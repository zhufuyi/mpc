package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zhufuyi/mpc/promConf"
)

func addCommand() *cobra.Command {
	var (
		resourceArg                         string
		fileFlag, jobNameFlag, jobValueFlag string
		valuesFlag                          []string
		keyValuesFlag                       = mapFlag{}
	)

	cmd := &cobra.Command{
		Use:   "add <resource>",
		Short: "Add job,targets,labels to prometheus configuration file",
		Long: `add job,targets,labels to prometheus configuration file.

Examples:
    # append new value to job'targets
    mpc add targets -f prometheus.yaml -n node_exporter -v 127.0.0.1:9100

    # append new kv to job'labels
    mpc add labels -f prometheus.yaml -n node_exporter -p foo=bar

    # add or replace job
    mpc add job -f prometheus.yaml -n node_exporter -d '
job_name: mysql_exporter
static_configs:
- targets:
  - 127.0.0.0:3306
  labels:
    foo: bar'
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
			// 执行job命令
			case Job:
				err := runJobAddCommand(&jobAddOptions{
					file:   fileFlag,
					values: jobValueFlag,
				})
				if err != nil {
					return err
				}

			// 执行targets命令
			case Targets:
				if err := checkJobName(jobNameFlag, "add"); err != nil {
					return err
				}
				if err := checkSliceValues(valuesFlag, "add"); err != nil {
					return err
				}
				err := runTargetsAddCommand(&targetsAddOptions{
					file:   fileFlag,
					name:   jobNameFlag,
					values: valuesFlag,
				})
				if err != nil {
					return err
				}

			// 执行targets命令
			case Labels:
				if err := checkJobName(jobNameFlag, "add"); err != nil {
					return err
				}
				if err := checkMapValues(keyValuesFlag, "add"); err != nil {
					return err
				}
				err := runLabelsAddCommand(&labelsAddOptions{
					file:      fileFlag,
					name:      jobNameFlag,
					keyValues: keyValuesFlag,
				})
				if err != nil {
					return err
				}

			default:
				return fmt.Errorf(`unknown resource name '%s'. use "mpc resources" for a complete list of supported resources.\n`, resourceArg)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&fileFlag, "file", "f", "", "prometheus configuration file, required, eg: prometheus.yaml")
	cmd.MarkFlagRequired("file")
	cmd.Flags().StringVarP(&jobNameFlag, "name", "n", "", "job name, required, eg: node_exporter")
	cmd.Flags().StringVarP(&jobValueFlag, "job-value", "d", "", "document value, if the resource is 'job', required, data format is yaml or json")
	cmd.Flags().StringSliceVarP(&valuesFlag, "targets-value", "v", nil, "if the resource is 'targets', required, data format is string, eg: 127.0.0.1:9100")
	cmd.Flags().VarP(&keyValuesFlag, "labels-value", "p", "key-value pairs, if the resource is 'labels', required, eg: foo=bar")

	return cmd
}

// ---------------------------------------------------------------------------------------

type jobAddOptions struct {
	file   string
	values string
}

func runJobAddCommand(options *jobAddOptions) error {
	data, err := readPrometheusConfigFile(options.file)
	if err != nil {
		return err
	}

	cy := promConf.NewConfigYaml(data)
	err = cy.AddJob([]byte(options.values))
	if err != nil {
		return err
	}

	return cy.Persistent(options.file)
}

type targetsAddOptions struct {
	file   string
	name   string
	values []string
}

func runTargetsAddCommand(options *targetsAddOptions) error {
	data, err := readPrometheusConfigFile(options.file)
	if err != nil {
		return err
	}

	cy := promConf.NewConfigYaml(data)
	err = cy.AddJobTargets(options.name, options.values)
	if err != nil {
		return err
	}

	return cy.Persistent(options.file)
}

type labelsAddOptions struct {
	file      string
	name      string
	keyValues mapFlag
}

func runLabelsAddCommand(options *labelsAddOptions) error {
	data, err := readPrometheusConfigFile(options.file)
	if err != nil {
		return err
	}

	cy := promConf.NewConfigYaml(data)
	err = cy.AddJobLabels(options.name, options.keyValues)
	if err != nil {
		return err
	}

	return cy.Persistent(options.file)
}

type mapFlag map[string]string

func (m mapFlag) String() string {
	return fmt.Sprintf("%v", map[string]string(m))
}

func (m mapFlag) Set(value string) error {
	split := strings.SplitN(value, "=", 2)
	if len(split) != 2 {
		return fmt.Errorf("value format is invalid, eg: env=dev")
	}

	// remove spaces
	m[strings.Trim(split[0], " ")] = strings.Trim(split[1], " ")

	return nil
}

func (m mapFlag) Type() string {
	return fmt.Sprintf("%T", m)
}

// ---------------------------------------------------------------------------------------

func checkJobName(jobName string, action string) error {
	if jobName == "" {
		return fmt.Errorf("you must specify the job_name of resource to %s. ", action)
	}

	return nil
}

func checkSliceValues(values []string, action string) error {
	if len(values) == 0 {
		return fmt.Errorf("you must specify the value of resource to %s. ", action)
	}

	return nil
}

func checkMapValues(values mapFlag, action string) error {
	if len(values) == 0 {
		return fmt.Errorf("you must specify the key-value of resource to %s", action)
	}

	return nil
}
