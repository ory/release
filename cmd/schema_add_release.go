package cmd

import (
	"github.com/spf13/cobra"
)

var serviceSchemaRefFormats = map[string]string{
	"kratos": "https://raw.githubusercontent.com/ory/kratos/%s/.schema/config.schema.json",
}

var schemaAddReleaseCmd = &cobra.Command{
	Use:   "add-release",
	Short: "Add a release to the version schema.",
	Long:  "Sth",
	Run: func(cmd *cobra.Command, args []string) {
		service, err := cmd.Flags().GetString("service")
		nerr(err)
		addVersionToMetaSchema(args[0], serviceSchemaRefFormats[service], args[1])
	},
}

func init() {
	schemaCmd.AddCommand(schemaAddReleaseCmd)

	schemaAddReleaseCmd.Flags().StringP("service", "s", "", "Set the service")
}
