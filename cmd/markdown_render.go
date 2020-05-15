package cmd

import (
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"
)

// renderCmd represents the render command
var renderCmd = &cobra.Command{
	Use:   "render <file.md>",
	Args:  cobra.ExactArgs(1),
	Short: "Renders a markdown file",
	Run: func(cmd *cobra.Command, args []string) {
		changelogRaw, err := ioutil.ReadFile(args[0])
		nerr(err)

		fmt.Println(renderMarkdown(changelogRaw))
	},
}

func init() {
	markdownCmd.AddCommand(renderCmd)
}
