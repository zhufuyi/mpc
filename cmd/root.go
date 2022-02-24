package cmd

import (
	"github.com/spf13/cobra"
)

// Version 命令版本号
const Version = "0.0.1"

// NewRootCMD 命令入口
func NewRootCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "mpc",
		Long:          "manage prometheus configuration, add,delete,update job",
		SilenceErrors: true,
		SilenceUsage:  true,
		Version:       Version,
	}

	cmd.AddCommand(
		getCommand(),
		addCommand(),
		deleteCommand(),
		replaceCommand(),
		resourcesCommand(),
		reloadCommand(),
		execCommand(),
		execsCommand(),
	)
	return cmd
}
