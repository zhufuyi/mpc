package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zhufuyi/mpc/promConf"
)

func deleteCommand() *cobra.Command {
	var (
		resourceArg           string
		fileFlag, jobNameFlag string
		valuesFlag, keysFlag  []string
	)

	cmd := &cobra.Command{
		Use:   "delete <resource>",
		Short: "Delete job,targets,labels in prometheus configuration file",
		Long: `delete job,targets,labels in prometheus configuration file.

Examples:
    # delete job in prometheus configuration file
    mpc delete job -f prometheus.yaml -n node_exporter

    # delete element in job'targets
    mpc delete targets -f prometheus.yaml -n node_exporter -v 127.0.0.1:9100

    # delete key in  job'labels
    mpc delete labels -f prometheus.yaml -n node_exporter -k foo
`,
		SilenceErrors: true,
		SilenceUsage:  true,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New(`you must specify the type of resource to add, use "mpc resources" for a complete list of supported resources.\n`)
			}
			resourceArg = args[0]
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {

			switch resourceArg {
			// 执行job命令
			case Job:
				if err := checkJobName(jobNameFlag, "delete"); err != nil {
					return err
				}
				err := runJobDelCommand(&jobDelOptions{
					file: fileFlag,
					name: jobNameFlag,
				})
				if err != nil {
					return err
				}

			// 执行targets命令
			case Targets:
				if err := checkJobName(jobNameFlag, "delete"); err != nil {
					return err
				}
				if err := checkSliceValues(valuesFlag, "delete"); err != nil {
					return err
				}
				err := runTargetsDelCommand(&targetsDelOptions{
					file:   fileFlag,
					name:   jobNameFlag,
					values: valuesFlag,
				})
				if err != nil {
					return err
				}

			// 执行targets命令
			case Labels:
				if err := checkJobName(jobNameFlag, "delete"); err != nil {
					return err
				}
				if err := checkSliceValues(keysFlag, "delete"); err != nil {
					return err
				}
				err := runLabelsDelCommand(&labelsDelOptions{
					file: fileFlag,
					name: jobNameFlag,
					keys: keysFlag,
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
	cmd.Flags().StringSliceVarP(&valuesFlag, "values", "v", nil, "if the resource is 'targets', required, eg: 127.0.0.1:9100")
	cmd.Flags().StringSliceVarP(&keysFlag, "keys", "k", nil, "if the resource is 'labels', required, eg: foo")

	return cmd
}

// ---------------------------------------------------------------------------------------

type jobDelOptions struct {
	file string
	name string
}

func runJobDelCommand(options *jobDelOptions) error {
	data, err := readPrometheusConfigFile(options.file)
	if err != nil {
		return err
	}

	cy := promConf.NewConfigYaml(data)
	err = cy.DelJob(options.name)
	if err != nil {
		return err
	}

	return cy.Persistent(options.file)
}

type targetsDelOptions struct {
	file   string
	name   string
	values []string
}

func runTargetsDelCommand(options *targetsDelOptions) error {
	data, err := readPrometheusConfigFile(options.file)
	if err != nil {
		return err
	}

	cy := promConf.NewConfigYaml(data)
	err = cy.DelJobTargets(options.name, options.values)
	if err != nil {
		return err
	}

	return cy.Persistent(options.file)
}

type labelsDelOptions struct {
	file string
	name string
	keys []string
}

func runLabelsDelCommand(options *labelsDelOptions) error {
	data, err := readPrometheusConfigFile(options.file)
	if err != nil {
		return err
	}

	cy := promConf.NewConfigYaml(data)
	err = cy.DelJobLabels(options.name, options.keys)
	if err != nil {
		return err
	}

	return cy.Persistent(options.file)
}
