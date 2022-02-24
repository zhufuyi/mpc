package gssh

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	// ExecSuccess 执行成功标记字符
	ExecSuccess = "successful execution, exit 0\n"

	// Separator 分隔符号
	Separator = "\n\n"
)

// RemoteServerInfo 远程服务器信息
type RemoteServerInfo struct {
	Host     string
	Port     int
	User     string
	Password string
}

func (r *RemoteServerInfo) String() string {
	return fmt.Sprintf("host=%s, port=%d, user=%s", r.Host, r.Port, r.User)
}

// CheckConnect 检查是否可以连接到远程服务器
func (r *RemoteServerInfo) CheckConnect() error {
	client, err := NewPwdSSHClient(r.Host, r.Port, r.User, r.Password)
	if err != nil {
		return err
	}
	defer client.Close()
	time.Sleep(time.Millisecond * 100)
	return nil
}

// FileParams 文件参数
type FileParams struct {
	// 本地文件参数
	ShellFile      string `json:"shellFile"`      // 脚本文件，执行程序入口
	CompressedFile string `json:"compressedFile"` // zip或tar.gz压缩文件

	// 目标服务器路径
	UploadPath string `json:"uploadPath"` // 上传文件到目标服务器的文件路径，默认路径为/tmp/upload
}

// 获取所有准备上传到远程服务器文件，包括md文件
func (f *FileParams) getUploadFiles() ([]string, error) {
	uploadFiles := []string{}

	err := CRLF2LF(f.ShellFile)
	if err != nil {
		return nil, err
	}
	md5ShellFile, err := GenMd5File(f.ShellFile)
	if err != nil {
		return nil, err
	}
	uploadFiles = append(uploadFiles, f.ShellFile, md5ShellFile)

	if f.CompressedFile != "" {
		md5File, err := GenMd5File(f.CompressedFile)
		if err != nil {
			return nil, err
		}
		uploadFiles = append(uploadFiles, f.CompressedFile, md5File)
	}

	return uploadFiles, nil
}

// 删除新生成的md5文件
func deleteMd5Files(uploadFiles []string) {
	for _, file := range uploadFiles {
		if strings.TrimRight(file, ".md5") != file {
			os.RemoveAll(file)
		}
	}
}

// 生成执行脚本命令
func (f FileParams) generateCmd() string {
	shellFilename := filepath.Base(f.ShellFile)
	compressedFilename := filepath.Base(f.CompressedFile)

	return fmt.Sprintf("bash %s/%s %s %s", f.UploadPath, shellFilename, f.UploadPath, compressedFilename)
}

// ExecShell 在远程服务器执行shell脚本，支持多个服务器
func ExecShell(ctx context.Context, servers []*RemoteServerInfo, fileParams *FileParams, outMsg chan string) {
	defer close(outMsg)

	uploadFiles, err := fileParams.getUploadFiles()
	defer deleteMd5Files(uploadFiles)
	if err != nil {
		outMsg <- fmt.Sprintf("getUploadFiles error, %v\n", err)
		return
	}

	if fileParams.UploadPath == "" {
		fileParams.UploadPath = "/tmp/upload"
	}
	outMsg <- Separator

	for _, server := range servers {
		// 连接远程服务器
		outMsg <- fmt.Sprintf("connecting remote server %s\n", server.Host)
		client, err := NewPwdSSHClient(server.Host, server.Port, server.User, server.Password)
		if err != nil {
			outMsg <- fmt.Sprintf("NewPwdSSHClient error, %v, %s\n", err, server.String())
			return
		}
		defer client.Close()

		// 发送文件到远程服务器
		for _, localFile := range uploadFiles {
			fi, err := os.Stat(localFile)
			if err != nil {
				fmt.Println("Stat error", err, localFile)
				continue
			}
			outMsg <- fmt.Sprintf("sending file '%s' to remote server %s, size=%dBytes ......\n", filepath.Base(localFile), server.Host, fi.Size())
			err = client.SendFile(ctx, localFile, fileParams.UploadPath)
			if err != nil {
				outMsg <- fmt.Sprintf("SendFile() error, err=%v, localFile=%s, remotePath=%s\n", err, filepath.Base(localFile), fileParams.UploadPath)
				return
			}
		}

		// 执行脚本
		cmd := fileParams.generateCmd()
		outMsg <- fmt.Sprintf("running command in remote server %s\n", server.Host)
		result := &Result{}
		client.Exec(ctx, cmd, result)
		for msg := range result.StdOut {
			outMsg <- msg
		}
		if result.Err != nil {
			outMsg <- fmt.Sprintf("Exec error, %v, cmd=%s\n", result.Err, cmd)
			return
		}

		outMsg <- Separator
	}

	outMsg <- ExecSuccess
}

//CRLF2LF 把windows文本\r\n转为unix的\n
func CRLF2LF(file string) error {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	crlf := []byte{'\r', '\n'}
	lf := []byte{'\n'}

	if !bytes.Contains(data, crlf) {
		return nil
	}

	data = bytes.ReplaceAll(data, crlf, lf)
	err = ioutil.WriteFile(file, data, 0777)
	if err != nil {
		return err
	}

	return nil
}
