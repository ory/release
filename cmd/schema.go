package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var schemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Schema related stuff",
	Long:  "Sth",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("nothing to do here...")
	},
}

func init() {
	rootCmd.AddCommand(schemaCmd)
}
