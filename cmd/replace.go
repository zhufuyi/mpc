package cmd

import (
	"fmt"
	"mpc/promConf"

	"github.com/spf13/cobra"
)

func replaceCommand() *cobra.Command {
	var (
		resourceArg           string
		fileFlag, jobNameFlag string
		valuesFlag            []string
		keyValuesFlag         = mapFlag{}
	)

	cmd := &cobra.Command{
		Use:   "replace <resource>",
		Short: "Replace job,targets,labels to prometheus configuration file",
		Long: `replace job,targets,labels to prometheus configuration file.

Examples:
    mpc replace targets -f prometheus.yaml -n node_exporter -v 127.0.0.1:9100

    mpc replace labels -f prometheus.yaml -n node_exporter -p foo=bar
`,
		SilenceErrors: true,
		SilenceUsage:  true,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf(`you must specify the type of resource to add, Use "mpc resources" for a complete list of supported resources.\n`)
			}
			resourceArg = args[0]
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {

			switch resourceArg {
			// 执行targets命令
			case Targets:
				if err := checkJobName(jobNameFlag, "replace"); err != nil {
					return err
				}
				if err := checkSliceValues(valuesFlag, "replace"); err != nil {
					return err
				}
				err := runTargetsReplaceCommand(&targetsReplaceOptions{
					file:   fileFlag,
					name:   jobNameFlag,
					values: valuesFlag,
				})
				if err != nil {
					return err
				}

			// 执行targets命令
			case Labels:
				if err := checkJobName(jobNameFlag, "replace"); err != nil {
					return err
				}
				if err := checkMapValues(keyValuesFlag, "replace"); err != nil {
					return err
				}
				err := runLabelsReplaceCommand(&labelsReplaceOptions{
					file:      fileFlag,
					name:      jobNameFlag,
					keyValues: keyValuesFlag,
				})
				if err != nil {
					return err
				}

			default:
				return fmt.Errorf("unknown resource name '%s'. Use \"mpc resources\" for a complete list of supported resources.\n", resourceArg)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&fileFlag, "file", "f", "", "prometheus configuration file, required, eg: prometheus.yaml")
	cmd.MarkFlagRequired("file")
	cmd.Flags().StringVarP(&jobNameFlag, "name", "n", "", "job name, required, eg: node_exporter")
	cmd.MarkFlagRequired("name")
	cmd.Flags().StringSliceVarP(&valuesFlag, "targets-value", "v", nil, "if the resource is 'targets', required, data format is string, eg: 127.0.0.1:9100")
	cmd.Flags().VarP(&keyValuesFlag, "labels-value", "p", "key-value pairs, if the resource is 'labels', required, eg: foo=bar")

	return cmd
}

// ---------------------------------------------------------------------------------------

type targetsReplaceOptions struct {
	file   string
	name   string
	values []string
}

func runTargetsReplaceCommand(options *targetsReplaceOptions) error {
	data, err := readPrometheusConfigFile(options.file)
	if err != nil {
		return err
	}

	cy := promConf.NewConfigYaml(data)
	err = cy.ReplaceJobTargets(options.name, options.values)
	if err != nil {
		return err
	}

	return cy.Persistent(options.file)
}

type labelsReplaceOptions struct {
	file      string
	name      string
	keyValues mapFlag
}

func runLabelsReplaceCommand(options *labelsReplaceOptions) error {
	data, err := readPrometheusConfigFile(options.file)
	if err != nil {
		return err
	}

	cy := promConf.NewConfigYaml(data)
	err = cy.ReplaceJobLabels(options.name, options.keyValues)
	if err != nil {
		return err
	}

	return cy.Persistent(options.file)
}
