package renderer

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/larsenclose/oci-tf-bootstrap/internal/discovery"
)

var nonAlphaNum = regexp.MustCompile(`[^a-z0-9]+`)
var leadingDigit = regexp.MustCompile(`^[0-9]`)

func toTFName(s string) string {
	s = strings.ToLower(s)
	s = nonAlphaNum.ReplaceAllString(s, "_")
	s = strings.Trim(s, "_")
	if leadingDigit.MatchString(s) {
		s = "n" + s
	}
	if s == "" {
		s = "unnamed"
	}
	return s
}

var providerTmpl = `terraform {
  required_providers {
    oci = {
      source  = "oracle/oci"
      version = ">= 5.0"
    }
  }
}

provider "oci" {
  region = "{{ .Region }}"
}
`

func writeProvider(result *discovery.Result, outputDir string) (err error) {
	tmpl := template.Must(template.New("provider").Parse(providerTmpl))
	f, err := os.Create(filepath.Join(outputDir, "provider.tf")) // #nosec G304 -- outputDir is user-specified CLI flag
	if err != nil {
		return err
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	data := map[string]string{
		"Region": result.Tenancy.HomeRegion,
	}
	return tmpl.Execute(f, data)
}
