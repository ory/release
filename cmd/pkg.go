package cmd

import (
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
	"github.com/markbates/pkger/pkging"
	"github.com/ory/gochimp3"
	"github.com/ory/x/httpx"
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
		parser.NoIntraEmphasis | parser.Tables | parser.FencedCode | parser.NoEmptyLineBeforeBlock |
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

func campaignID() string {
	return fmt.Sprintf("%s-%s-%s",
		getenv("CIRCLE_PROJECT_REPONAME"),
		getenv("CIRCLE_SHA1"),
		getenv("CIRCLE_TAG"),
	)
}
