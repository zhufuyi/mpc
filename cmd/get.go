package cmd

import (
	"fmt"
	"io/ioutil"
	"mpc/promConf"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func getCommand() *cobra.Command {
	var (
		resourceArg           string
		fileFlag, jobNameFlag string
	)

	cmd := &cobra.Command{
		Use:   "get <resource>",
		Short: "Show job,targets,labels from prometheus configuration file",
		Long: `show job,targets,labels from prometheus configuration file.

Examples:
    mpc get job -f prometheus.yaml -n node_exporter

    mpc get targets -f prometheus.yaml -n node_exporter

    mpc get labels -f prometheus.yaml -n node_exporter
`,
		SilenceErrors: true,
		SilenceUsage:  true,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf(`you must specify the type of resource to get. use "mpc resources" for a complete list of supported resources.\n`)
			}
			resourceArg = args[0]
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {

			switch resourceArg {
			case Job:
				job, err := runJobGetCommand(&jobGetOptions{
					file: fileFlag,
					name: jobNameFlag,
				})
				if err != nil {
					return err
				}
				fmt.Println(string(job))

			case Targets:
				targets, err := runTargetsGetCommand(&targetsGetOptions{
					file: fileFlag,
					name: jobNameFlag,
				})
				if err != nil {
					return err
				}
				fmt.Println(targets)

			case Labels:
				labels, err := runLabelsGetCommand(&labelsGetOptions{
					file: fileFlag,
					name: jobNameFlag,
				})
				if err != nil {
					if strings.Contains(err.Error(), "no value found") {
						fmt.Println(labels)
						return nil
					}
					return err
				}
				fmt.Println(labels)

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

	return cmd
}

// ---------------------------------------------------------------------------------------

type jobGetOptions struct {
	file string
	name string
}

func runJobGetCommand(options *jobGetOptions) ([]byte, error) {
	data, err := readPrometheusConfigFile(options.file)
	if err != nil {
		return nil, err
	}

	cy := promConf.NewConfigYaml(data)
	return cy.GetJob(options.name)
}

type targetsGetOptions struct {
	file string
	name string
}

func runTargetsGetCommand(options *targetsGetOptions) ([]string, error) {
	data, err := readPrometheusConfigFile(options.file)
	if err != nil {
		return nil, err
	}

	cy := promConf.NewConfigYaml(data)
	return cy.GetJobTargets(options.name)
}

type labelsGetOptions struct {
	file string
	name string
}

func runLabelsGetCommand(options *labelsGetOptions) (map[string]string, error) {
	data, err := readPrometheusConfigFile(options.file)
	if err != nil {
		return nil, err
	}

	cy := promConf.NewConfigYaml(data)
	return cy.GetJobLabels(options.name)
}

// ---------------------------------------------------------------------------------------

func readPrometheusConfigFile(file string) ([]byte, error) {
	_, err := os.Stat(file)
	if err != nil {
		return nil, err
	}

	return ioutil.ReadFile(file)
}
