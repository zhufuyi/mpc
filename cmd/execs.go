package cmd

import "github.com/spf13/cobra"

func execsCommand() *cobra.Command {
	var (
		serversListFlag, execScriptFlag, installFileFlag, uploadPathFlag string
	)

	cmd := &cobra.Command{
		Use:   "execs",
		Short: "Install and run service to multiple remote servers",
		Long: `install and run service to multiple remote servers.

Examples:
    mpc execs -j remote_servers.json -e node_exporter_install.sh -f node_exporter-1.3.1.linux-amd64.tar.gz
`,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {

			err := runExecCommand(&execGetOptions{
				serversList: serversListFlag,
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

	cmd.Flags().StringVarP(&serversListFlag, "servers-list", "j", "", `server address list file, data format is json, file content example:
  [
    {
      "host": "192.168.1.11",
      "port": 22,
      "user": "root",
      "password": "1234"
    }
  ]`)
	cmd.Flags().StringVarP(&execScriptFlag, "execute-script", "e", "", "execute script file, written by users themselves, required")
	cmd.MarkFlagRequired("execute-script")
	cmd.Flags().StringVarP(&installFileFlag, "install-file", "f", "", "install file, format is '.zip' or '.tar.gz'")
	cmd.Flags().StringVarP(&uploadPathFlag, "upload-path", "d", "/tmp/upload", "specify the path to upload files to the remote server")

	return cmd
}
