package cli

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"text/template"

	"github.com/insomniacslk/sre/pkg/ansi"
	"github.com/insomniacslk/sre/pkg/config"

	"github.com/fatih/color"
	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewOmgCmd(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "omg",
		Short: "First-responder tool",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			logrus.Debugf("Running omg command")
			templateFile := cfg.Omg.Template
			templateBytes, err := os.ReadFile(templateFile)
			if err != nil {
				logrus.Fatalf("Failed to read template file: %v", err)
			}
			funcMap := template.FuncMap{
				"url": func(text, url string) string {
					return ansi.ToURL(text, url)
				},
				"ruler": func() string {
					return strings.Repeat("=", 80)
				},
				"image": func(img []byte) (string, error) {
					return ansi.ToTerminalImage(img)
				},
				"cat": func(file string) ([]byte, error) {
					file, err := homedir.Expand(file)
					if err != nil {
						return nil, err
					}
					buf, err := os.ReadFile(file)
					if err != nil {
						return nil, err
					}
					return buf, nil
				},
				"fetch": func(url string) ([]byte, error) {
					resp, err := http.Get(url)
					if err != nil {
						return nil, err
					}
					defer func() {
						if err := resp.Body.Close(); err != nil {
							logrus.Warningf("Failed to close response body: %v", err)
						}
					}()
					if resp.StatusCode != 200 {
						return nil, fmt.Errorf("expected 200 OK, got %s", resp.Status)
					}
					buf, err := io.ReadAll(resp.Body)
					if err != nil {
						return nil, err
					}
					return buf, nil
				},
				"bold": func(text string) string {
					return ansi.Bold(text)
				},
				"italic": func(text string) string {
					return ansi.Italic(text)
				},
				"red": func(text string) string {
					return color.RedString(text)
				},
				"green": func(text string) string {
					return color.GreenString(text)
				},
				"blue": func(text string) string {
					return color.BlueString(text)
				},
			}
			tmpl, err := template.New("omg").Funcs(funcMap).Parse(string(templateBytes))
			if err != nil {
				logrus.Fatalf("Failed to parse template: %v", err)
			}
			err = tmpl.Execute(os.Stdout, nil)
			if err != nil {
				logrus.Fatalf("Failed to execute template: %v", err)
			}
		},
	}
}
