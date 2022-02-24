package gssh

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// SSHClient 远程服务器信息
type SSHClient struct {
	host     string
	user     string
	password string
	sshKey   []byte
	port     int
	client   *ssh.Client  // ssh客户端端
	sftpCli  *sftp.Client // sftp客户端，依赖ssh客户端
}

// NewPwdSSHClient 连接远程服务器，密码方式
func NewPwdSSHClient(host string, port int, user string, password string) (*SSHClient, error) {
	cli := &SSHClient{
		host:     host,
		user:     user,
		password: password,
		port:     port,
	}

	err := cli.connect("pwd")
	if err != nil {
		return nil, err
	}

	return cli, nil
}

// NewKeySSHClient 连接远程服务器，key方式
func NewKeySSHClient(host string, port int, user string, sshKey []byte) (*SSHClient, error) {
	cli := &SSHClient{
		host:   host,
		user:   user,
		sshKey: sshKey,
		port:   port,
	}

	err := cli.connect("key")
	if err != nil {
		return nil, err
	}

	return cli, nil
}

// Exec 执行命令，实时信息返回在result对象中
func (s *SSHClient) Exec(ctx context.Context, cmd string, result *Result) {
	exit := make(chan struct{})
	initResult(result)

	session, err := s.client.NewSession()
	if err != nil {
		result.setErrMsg(err)
		close(result.StdOut)
		return
	}

	go func() {
		defer func() {
			defer close(result.StdOut) // 执行完毕，关闭通道
			defer session.Close()
		}()

		execCmd(session, cmd, result, exit)
	}()

	go func() {
		defer func() {
			session.Close()
		}()

		for {
			select {
			case <-ctx.Done():
				session.Signal(ssh.SIGKILL)
				return
			case <-exit: // 退出协程，防止泄露
				return
			}
		}
	}()
}

// Execs 批量执行命令，按顺序执行，如果前面命令执行失败，不会执行后面命令，实时信息返回在result对象中
func (s *SSHClient) Execs(ctx context.Context, cmds []string, result *Result) {
	cmd := strings.Join(cmds, " && ")
	s.Exec(ctx, cmd, result)
}

func (s *SSHClient) connect(kind string) error {
	sshConfig := &ssh.ClientConfig{
		Timeout:         15 * time.Second,
		User:            s.user,
		HostKeyCallback: ssh.HostKeyCallback(func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil }),
	}

	switch kind {
	case "pwd":
		sshConfig.Auth = []ssh.AuthMethod{
			ssh.Password(s.password),
		}

	case "key":
		if len(s.sshKey) == 0 {
			return errors.New("ssh key is empty")
		}
		// 创建秘钥签名
		signer, err := ssh.ParsePrivateKey(s.sshKey)
		if err != nil {
			return err
		}
		// 配置秘钥登录
		sshConfig.Auth = []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		}
	}

	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	conn, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return err
	}

	s.client = &ssh.Client{Conn: conn}

	return nil
}

// CreateSftp 创建sftp会话
func (s *SSHClient) CreateSftp() error {
	sftpCli, err := sftp.NewClient(s.client)
	if err != nil {
		return err
	}

	s.sftpCli = sftpCli
	return nil
}

// Close 关闭ssh连接
func (s *SSHClient) Close() error {
	if s == nil {
		return nil
	}
	if s.sftpCli != nil {
		if err := s.sftpCli.Close(); err != nil {
			return err
		}
	}
	if s.client != nil {
		if err := s.client.Close(); err != nil {
			return err
		}
	}
	return nil
}

// CloseSftp 关闭sftp连接
func (s *SSHClient) CloseSftp() error {
	if s == nil {
		return nil
	}
	if s.sftpCli != nil {
		if err := s.sftpCli.Close(); err != nil {
			return err
		}
	}
	return nil
}

func getRemoteFile(remoteDir string, localFile string) string {
	remoteDir = strings.TrimRight(remoteDir, "/")
	return remoteDir + "/" + path.Base(localFile)
}

// SendFile 发送文件到远程服务器
func (s *SSHClient) SendFile(ctx context.Context, localFile string, remoteDir string) error {
	if s.sftpCli == nil {
		if err := s.CreateSftp(); err != nil {
			return err
		}
	}

	srcFile, err := os.Open(localFile)
	if err != nil {
		return fmt.Errorf("Open() local file %s error, err=%v", localFile, err)
	}
	defer srcFile.Close()

	// 在远程创建目录，mkdir -p
	err = s.sftpCli.MkdirAll(remoteDir)
	if err != nil {
		return err
	}
	remoteFile := getRemoteFile(remoteDir, localFile)
	dstFile, err := s.sftpCli.Create(remoteFile)
	if err != nil {
		return fmt.Errorf("Create() remove file %s error, err=%v", remoteFile, err)
	}
	defer dstFile.Close()

	// 写入内容
	bufSize := 40960 // 一次读取字节数
	readBuf := bufio.NewReader(srcFile)
	buf := make([]byte, bufSize)

	writerBuf := bufio.NewWriterSize(dstFile, bufSize)
	isEOF := false

SENDEND:
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("cancel or time out")
		default:
			// 读取
			n, err := readBuf.Read(buf)
			if err != nil {
				if err == io.EOF {
					isEOF = true
				} else {
					return err
				}
			}

			// 写入
			_, err = writerBuf.Write(buf[:n])
			if err != nil {
				return err
			}
			if isEOF {
				break SENDEND
			}
		}
	}

	if err := writerBuf.Flush(); err != nil {
		return err
	}

	return nil
}

// SendContent 发送文件内容到远程服务器
func (s *SSHClient) SendContent(ctx context.Context, filename string, content []byte, remoteDir string) error {
	if s.sftpCli == nil {
		if err := s.CreateSftp(); err != nil {
			return err
		}
	}

	// 在远程创建目录，mkdir -p
	err := s.sftpCli.MkdirAll(remoteDir)
	if err != nil {
		return err
	}
	remoteFile := getRemoteFile(remoteDir, filename)
	dstFile, err := s.sftpCli.Create(remoteFile)
	if err != nil {
		return fmt.Errorf("Create() remove file %s error, err=%v", remoteFile, err)
	}
	defer dstFile.Close()

	// 写入内容
	bufSize := 40960 // 一次读取字节数
	srcFile := bytes.NewReader(content)
	readBuf := bufio.NewReader(srcFile)
	buf := make([]byte, bufSize)

	writerBuf := bufio.NewWriterSize(dstFile, bufSize)
	isEOF := false

SENDEND:
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("cancel or time out")
		default:
			// 读取
			n, err := readBuf.Read(buf)
			if err != nil {
				if err == io.EOF {
					isEOF = true
				} else {
					return err
				}
			}

			// 写入
			_, err = writerBuf.Write(buf[:n])
			if err != nil {
				return err
			}
			if isEOF {
				break SENDEND
			}
		}
	}

	if err := writerBuf.Flush(); err != nil {
		return err
	}

	return nil
}

// Result 执行命令的结果
type Result struct {
	StdOut chan string
	Err    error
	rwMux  *sync.RWMutex
}

func (r *Result) setErrMsg(err error) {
	r.rwMux.Lock()
	defer r.rwMux.Unlock()
	r.Err = fmt.Errorf("%s", err)
}

func initResult(result *Result) {
	if result == nil {
		result = &Result{StdOut: make(chan string), Err: error(nil)}
		return
	}
	result.rwMux = &sync.RWMutex{}
	result.StdOut = make(chan string)
	result.Err = error(nil)
}

func execCmd(session *ssh.Session, cmd string, result *Result, exit chan struct{}) {
	defer close(exit)

	result.StdOut <- cmd + "\n"

	stdout, err := session.StdoutPipe()
	if err != nil {
		result.Err = fmt.Errorf("stdout error, err = %s", err.Error())
		return
	}

	stderr, err := session.StderrPipe()
	if err != nil {
		result.Err = fmt.Errorf("stderr error, err = %s", err.Error())
		return
	}

	err = session.Start(cmd)
	if err != nil {
		result.Err = fmt.Errorf("session start error, err = %s", err.Error())
		return
	}

	reader := bufio.NewReader(stdout)
	// 实时读取每行内容
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			// 判断是否已经读取完毕
			if err.Error() == io.EOF.Error() {
				break
			}

			result.Err = fmt.Errorf("stdout error, err = %s", err.Error())
			break
		}
		result.StdOut <- line
	}

	// 捕获错误日志
	bytesErr, err := ioutil.ReadAll(stderr)
	if err != nil {
		result.Err = fmt.Errorf("read stderr error, err = %s", err.Error())
		return
	}
	if len(bytesErr) != 0 {
		result.Err = fmt.Errorf("%s", bytesErr)
		return
	}

	err = session.Wait()
	if err != nil {
		result.Err = err
	}
}
