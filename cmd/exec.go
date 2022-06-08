package cmd

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/zhufuyi/mpc/gssh"

	"github.com/spf13/cobra"
)

func execCommand() *cobra.Command {
	var (
		userFlag, passwordFlag, hostFlag, execScriptFlag, installFileFlag, uploadPathFlag string
		portFlag                                                                          int
	)

	cmd := &cobra.Command{
		Use:   "exec",
		Short: "Install and run service to one remote server",
		Long: `install and run service to one remote server.

Examples:
    mpc exec -u root -p 123456 -H 192.168.1.10 -P 22 -e node_exporter_install.sh -f node_exporter-1.3.1.linux-amd64.tar.gz
`,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {

			err := runExecCommand(&execGetOptions{
				user:        userFlag,
				password:    passwordFlag,
				host:        hostFlag,
				port:        portFlag,
				execScript:  execScriptFlag,
				installFile: installFileFlag,
				UploadPath:  uploadPathFlag,
			})
			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&userFlag, "user", "u", "", "remote server user name")
	cmd.MarkFlagRequired("user")
	cmd.Flags().StringVarP(&passwordFlag, "password", "p", "", "remote server password")
	cmd.MarkFlagRequired("password")
	cmd.Flags().StringVarP(&hostFlag, "host", "H", "", "remote server host")
	cmd.MarkFlagRequired("host")
	cmd.Flags().IntVarP(&portFlag, "port", "P", 22, "remote server port")
	cmd.Flags().StringVarP(&execScriptFlag, "execute-script", "e", "", "execute script file, written by users themselves, required")
	cmd.MarkFlagRequired("execute-script")
	cmd.Flags().StringVarP(&installFileFlag, "install-file", "f", "", "install file, format is '.zip' or '.tar.gz'")
	cmd.Flags().StringVarP(&uploadPathFlag, "upload-path", "d", "/tmp/upload", "specify the path to upload files to the remote server")

	return cmd
}

// ----------------------------------------------------------------------------------------

type execGetOptions struct {
	user        string
	password    string
	host        string
	port        int
	serversList string
	execScript  string
	installFile string
	UploadPath  string
}

func runExecCommand(options *execGetOptions) error {
	servers := []*gssh.RemoteServerInfo{}

	// 优先使用文件列表
	if options.serversList != "" {
		data, err := ioutil.ReadFile(options.serversList)
		if err != nil {
			fmt.Printf("ReadFile error, %v\n", err)
			return err
		}
		rsis := []*gssh.RemoteServerInfo{}
		err = jsoniter.Unmarshal(data, &rsis)
		if err != nil {
			fmt.Printf("Unmarshal error, %v\n", err)
			return err
		}
		if len(rsis) == 0 {
			return errors.New("remote servers is empty, please set the flag 'servers-list' json file")
		}
		servers = rsis
	} else {
		servers = []*gssh.RemoteServerInfo{
			{
				Host:     options.host,
				Port:     options.port,
				User:     options.user,
				Password: options.password,
			},
		}
	}

	ctx, _ := context.WithTimeout(context.Background(), time.Minute*5)
	fileParams := &gssh.FileParams{
		ShellFile:      options.execScript,
		CompressedFile: options.installFile,
		UploadPath:     options.UploadPath,
	}

	outMsg := make(chan string)

	go gssh.ExecShell(ctx, servers, fileParams, outMsg)
	var msg string
	for msg = range outMsg {
		fmt.Printf(msg)
	}
	if msg != gssh.ExecSuccess {
		return errors.New("execute failed")
	}

	return nil
}
