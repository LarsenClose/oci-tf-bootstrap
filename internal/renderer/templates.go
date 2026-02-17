package renderer

import (
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/larsenclose/oci-tf-bootstrap/internal/discovery"
)

func toTFName(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, ".", "_")
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

func writeProvider(result *discovery.Result, outputDir string) error {
	tmpl := template.Must(template.New("provider").Parse(providerTmpl))
	f, err := os.Create(filepath.Join(outputDir, "provider.tf")) // #nosec G304 -- outputDir is user-specified CLI flag
	if err != nil {
		return err
	}
	defer f.Close()

	data := map[string]string{
		"Region": result.Tenancy.HomeRegion,
	}
	return tmpl.Execute(f, data)
}
