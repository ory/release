/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Version = "master"
	Date    = "undefined"
	Commit  = "undefined"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display this binary's version, build time and git hash of this build",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version:    %s\n", Version)
		fmt.Printf("Git Hash:   %s\n", Commit)
		fmt.Printf("Build Time: %s\n", Date)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
