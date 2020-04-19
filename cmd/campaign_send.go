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

	"github.com/ory/gochimp3"
	"github.com/ory/x/flagx"
	"github.com/spf13/cobra"
)

// campaignSendCmd represents the send command
var campaignSendCmd = &cobra.Command{
	Use: "send",
	Long: `Send a drafted campaign.

Example:

	$ MAILCHIMP_API_KEY=... \
		CIRCLE_SHA1=... \
		CIRCLE_TAG=... \ # This is set automatically in CircleCI Jobs
		CIRCLE_PROJECT_REPONAME=... \ # This is set automatically in CircleCI Jobs
		release campaign send
`,
	Run: func(cmd *cobra.Command, args []string) {
		chimpKey := getenv("MAILCHIMP_API_KEY")
		chimp := gochimp3.New(chimpKey)

		campaigns, err := chimp.GetCampaigns(&gochimp3.CampaignQueryParams{Status: "save"})
		nerr(err)

		for _, c := range campaigns.Campaigns {
			if c.Settings.Title == campaignID() {
				if flagx.MustGetBool(cmd, "dry") {
					fmt.Println("Skipping send because --dry was passed.")
					return
				}

				chimpCampaignSent, err := chimp.SendCampaign(c.ID, &gochimp3.SendCampaignRequest{
					CampaignId: c.ID,
				})
				nerr(err)

				if !chimpCampaignSent {
					fatalf("Unable to send MailChimp Campaign: %s", c.ID)
				}

				fmt.Println("Sent campaign!")
			}
		}
	},
}

func init() {
	campaignCmd.AddCommand(campaignSendCmd)

	campaignSendCmd.Flags().Bool("dry", false, "Do not ")
}
