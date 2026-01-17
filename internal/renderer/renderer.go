package renderer

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/larsenclose/oci-tf-bootstrap/internal/discovery"
)

// Options configures terraform output generation
type Options struct {
	AlwaysFree bool
}

func OutputJSON(result *discovery.Result, w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

func OutputTerraform(result *discovery.Result, outputDir string, opts Options) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	if err := writeProvider(result, outputDir); err != nil {
		return fmt.Errorf("provider.tf: %w", err)
	}
	if err := writeLocals(result, outputDir, opts); err != nil {
		return fmt.Errorf("locals.tf: %w", err)
	}
	if err := writeDataSources(result, outputDir); err != nil {
		return fmt.Errorf("data.tf: %w", err)
	}
	if err := writeInstanceExample(result, outputDir, opts); err != nil {
		return fmt.Errorf("instance_example.tf: %w", err)
	}
	if err := writeNetwork(result, outputDir, opts); err != nil {
		return fmt.Errorf("network.tf: %w", err)
	}
	return nil
}
