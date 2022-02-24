package cmd

import (
	"fmt"

	"mpc/promConf"

	"github.com/spf13/cobra"
)

func reloadCommand() *cobra.Command {
	var promURLFlag string

	cmd := &cobra.Command{
		Use:   "reload",
		Short: "Make the prometheus configuration effective",
		Long: `make the prometheus configuration effective

Examples:
    mpc reload -p http://127.0.0.1:9090/-/reload
`,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := promConf.ConfReload(promURLFlag)
			if err != nil {
				fmt.Println(err)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&promURLFlag, "promURL", "p", "http://127.0.0.1:9090/-/reload", "prometheus url")

	return cmd
}
