package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func resourcesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resources",
		Short: "List of supported resources",
		Long: `list of supported resources. 

Examples:
    mpc resources
`,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(string(ListResourceNames()))
			return nil
		},
	}

	return cmd
}

// --------------------------------------------------------------------------------------

const (
	// Job job资源
	Job = "job"
	// Targets job下targets资源
	Targets = "targets"
	// Labels job下labels资源
	Labels = "labels"
)

// 支持的资源名称列表
var resourceNames = []string{
	Job,
	Targets,
	Labels,
}

// ListResourceNames 资源名称列表
func ListResourceNames() []byte {
	content := []string{"resources list:\n\n"}
	for _, name := range resourceNames {
		content = append(content, name+"\n\n")
	}

	return []byte(strings.Join(content, ""))
}
