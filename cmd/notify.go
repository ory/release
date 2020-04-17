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
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/hanzoai/gochimp3"
	"github.com/markbates/pkger"
	"github.com/markbates/pkger/pkging"
	_ "github.com/ory/x/cmdx"
	"github.com/ory/x/flagx"
	"github.com/ory/x/httpx"
	"github.com/spf13/cobra"
)

func nerr(err error) {
	if err == nil {
		return
	}
	fatalf("An unexpected error occurred:\n\t%+v", err)
}

func fatalf(m string, args ...interface{}) {
	fmt.Printf(m, args...)
	fmt.Println()
	os.Exit(1)
}

func getenv(key string) (v string) {
	v = os.Getenv(key)
	if len(v) == 0 {
		fatalf("Environment variable " + key + " must be set.")
	}
	return
}

func readTemplate(file pkging.File, err error) *template.Template {
	nerr(err)
	defer file.Close()

	contents, err := ioutil.ReadAll(file)
	nerr(err)

	t, err := template.New(file.Name()).Parse(string(contents))
	nerr(err)
	return t
}

func renderMarkdown(source []byte) template.HTML {
	var markdownRenderer = html.NewRenderer(html.RendererOptions{Flags: html.CommonFlags | html.HrefTargetBlank})
	var markdownParser = parser.NewWithExtensions(
		parser.NoIntraEmphasis | parser.Tables | parser.FencedCode |
			parser.Autolink | parser.Strikethrough | parser.SpaceHeadings | parser.DefinitionLists)

	rendered := string(markdown.ToHTML(source, markdownParser, markdownRenderer))
	rendered = strings.ReplaceAll(rendered, "<p>", "")
	rendered = strings.ReplaceAll(rendered, "</p>", "<br>")
	return template.HTML(rendered)
}

func newMailchimpRequest(apiKey, path string, payload interface{}) {
	u := url.URL{}
	u.Scheme = "https"
	u.Host = fmt.Sprintf(gochimp3.URIFormat, gochimp3.DatacenterRegex.FindString(apiKey))
	u.Path = filepath.Join(gochimp3.Version, path)
	req, err := http.NewRequest("GET", u.String(), nil)
	nerr(err)
	req.SetBasicAuth("gochimp3", apiKey)
	client := httpx.NewResilientClientLatencyToleranceMedium(nil)
	res, err := client.Do(req)
	nerr(err)
	defer res.Body.Close()
	nerr(json.NewDecoder(res.Body).Decode(payload))
}

// notifyCmd represents the notify command
var notifyCmd = &cobra.Command{
	Use:   "notify list-id path/to/tag-message path/to/changelog.md",
	Args:  cobra.ExactArgs(3),
	Short: "Sends out a release notification via the Mailchimp Campaign / Newsletter API",
	Long: `TL;DR

	$ git tag -l --format='%(contents)' v0.0.103 > tag-message.txt
	$ # run changelog generator > changelog.md
	$ MAILCHIMP_API_KEY=... \
		CIRCLE_TAG=... \ # This is set automatically in CircleCI Jobs
		CIRCLE_PROJECT_REPONAME=... \ # This is set automatically in CircleCI Jobs
		release notify \
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

If you want to send only to a segment within that list, add the Segment ID as well:

	release notify --segment 1234 ...
`,
	Run: func(cmd *cobra.Command, args []string) {
		repoName := getenv("CIRCLE_PROJECT_NAME")
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
			fmt.Printf("%+v\n", segmentOptions)
		}

		chimpCampaign, err := chimp.CreateCampaign(&gochimp3.CampaignCreationRequest{
			Type: gochimp3.CAMPAIGN_TYPE_REGULAR,
			Recipients: gochimp3.CampaignCreationRecipients{
				ListId:         listID,
				SegmentOptions: segmentOptions,
			},
			Settings: gochimp3.CampaignCreationSettings{
				Title:        fmt.Sprintf("%s %s Release Announcement", projectName, tag),
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

		fmt.Printf("Created campaign: %s", chimpCampaign.ID)
		fmt.Println()

		chimpCampaignSent,err := chimp.SendCampaign(chimpCampaign.ID, &gochimp3.SendCampaignRequest{
			CampaignId: chimpCampaign.ID,
		})
		nerr(err)

		if !chimpCampaignSent{
			fatalf("Unable to send MailChimp Campaign: %s", chimpCampaign.ID)
		}

		fmt.Println("Sent campaign!")
	},
}

func init() {
	rootCmd.AddCommand(notifyCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// notifyCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	notifyCmd.Flags().Int("segment", 0, "The Mailchimp Segment ID")
}
