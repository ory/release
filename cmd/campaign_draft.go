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
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"strings"

	"github.com/markbates/pkger"
	"github.com/ory/gochimp3"
	_ "github.com/ory/x/cmdx"
	"github.com/ory/x/flagx"
	"github.com/spf13/cobra"
)

// campaignDraftCmd represents the notify command
var campaignDraftCmd = &cobra.Command{
	Use:   "draft list-id path/to/tag-message path/to/changelog.md",
	Args:  cobra.ExactArgs(3),
	Short: "Creates a draft release notification via the Mailchimp Campaign / Newsletter API",
	Long: `TL;DR

	$ git tag -l --format='%(contents)' v0.0.103 > tag-message.txt
	$ # run changelog generator > changelog.md
	$ MAILCHIMP_API_KEY=... \
		CIRCLE_SHA1=... \
		CIRCLE_TAG=... \ # This is set automatically in CircleCI Jobs
		CIRCLE_PROJECT_REPONAME=... \ # This is set automatically in CircleCI Jobs
		release campaign draft \
			--segment-id ... \ # optional - e.g. only to people interested in ORY Hydra
			list-id-1234123 \
			./tag-message.md \
			./changelog.md

To send out a release newsletter you need to specify an API Key for Mailchimp
(https://admin.mailchimp.com/account/api/) using the MAILCHIMP_API_KEY environment variable:

	export MAILCHIMP_API_KEY=...

Additionally, these CI environment variables are expected to be set as well:

	$CIRCLE_PROJECT_REPONAME (e.g. hydra)
	$CIRCLE_TAG (e.g. v1.4.5-beta.1)
	$CIRCLE_SHA1

If you want to send only to a segment within that list, add the Segment ID as well:

	release notify --segment 1234 ...
`,
	Run: func(cmd *cobra.Command, args []string) {
		repoName := getenv("CIRCLE_PROJECT_REPONAME")
		projectName := "ORY " + strings.Title(strings.ToLower(repoName))
		tag := getenv("CIRCLE_TAG")
		listID := args[0]
		tagMessagePath := args[1]
		changelogPath := args[2]

		tagMessageRaw, err := ioutil.ReadFile(tagMessagePath)
		nerr(err)
		changelogRaw, err := ioutil.ReadFile(changelogPath)
		nerr(err)

		changelog := renderMarkdown(changelogRaw)
		tagMessage := renderMarkdown(tagMessageRaw)

		var body bytes.Buffer
		nerr(readTemplate(pkger.Open("/view/mail-body.html")).Execute(&body, struct {
			Version     string
			GitTag      string
			ProjectName string
			RepoName    string
			Changelog   template.HTML
			Message     template.HTML
		}{
			Version:     tag,
			GitTag:      tag,
			ProjectName: projectName,
			RepoName:    repoName,
			Changelog:   changelog,
			Message:     tagMessage,
		}))

		chimpKey := getenv("MAILCHIMP_API_KEY")
		chimp := gochimp3.New(chimpKey)
		chimpTemplate, err := chimp.CreateTemplate(&gochimp3.TemplateCreationRequest{
			Name: fmt.Sprintf("%s %s Release Announcement", projectName, tag),
			Html: body.String(),
		})
		nerr(err)

		var segmentOptions *gochimp3.CampaignCreationSegmentOptions
		if segmentID := flagx.MustGetInt(cmd, "segment"); segmentID > 0 {
			var payload struct {
				Options *gochimp3.CampaignCreationSegmentOptions `json:"options"`
			}
			newMailchimpRequest(chimpKey, fmt.Sprintf("/lists/%s/segments/%d", listID, segmentID), &payload)
			segmentOptions = payload.Options
			segmentOptions.SavedSegmentId = segmentID
		}

		chimpCampaign, err := chimp.CreateCampaign(&gochimp3.CampaignCreationRequest{
			Type: gochimp3.CAMPAIGN_TYPE_REGULAR,
			Recipients: gochimp3.CampaignCreationRecipients{
				ListId:         listID,
				SegmentOptions: segmentOptions,
			},
			Settings: gochimp3.CampaignCreationSettings{
				Title:        campaignID(),
				SubjectLine:  fmt.Sprintf("%s %s has been released!", projectName, tag),
				FromName:     "ORY",
				ReplyTo:      "hi@ory.sh",
				Authenticate: true,
				FbComments:   false,
				TemplateId:   chimpTemplate.ID,
			},
			Tracking: gochimp3.CampaignTracking{
				Opens:      true,
				HtmlClicks: true,
				TextClicks: true,
			},
		})
		nerr(err)

		fmt.Printf(`Created campaign "%s" (%s)`, chimpCampaign.Settings.Title, chimpCampaign.ID)
		fmt.Println()

		fmt.Println("Campaign drafted")
	},
}

func init() {
	campaignCmd.AddCommand(campaignDraftCmd)
	campaignDraftCmd.Flags().Int("segment", 0, "The Mailchimp Segment ID")
}
