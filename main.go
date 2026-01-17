package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/larsenclose/oci-tf-bootstrap/internal/discovery"
	"github.com/larsenclose/oci-tf-bootstrap/internal/renderer"
)

// Version information - set via ldflags at build time
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var (
	profile     = flag.String("profile", "", "OCI config profile name (default: $OCI_CLI_PROFILE or DEFAULT)")
	configDir   = flag.String("config", "", "OCI config directory (default: $OCI_CLI_CONFIG_FILE directory or ~/.oci)")
	configFile  = flag.String("config-file", "", "OCI config file path (default: $OCI_CLI_CONFIG_FILE or ~/.oci/config)")
	outputDir   = flag.String("output", "./terraform", "Output directory for generated TF files")
	region      = flag.String("region", "", "Override region (default: from config)")
	jsonOut     = flag.Bool("json", false, "Output raw discovery as JSON instead of TF")
	alwaysFree  = flag.Bool("always-free", false, "Filter output to always-free tier eligible resources only")
	showVersion = flag.Bool("version", false, "Print version information and exit")
)

// resolveConfigPath determines the OCI config file path from flags and environment variables.
// Priority: --config-file > --config > $OCI_CLI_CONFIG_FILE > ~/.oci/config
func resolveConfigPath() (configPath string, configDirectory string) {
	// --config-file takes highest priority
	if *configFile != "" {
		configPath = *configFile
		configDirectory = filepath.Dir(configPath)
		return
	}

	// --config (directory) takes second priority
	if *configDir != "" {
		configDirectory = *configDir
		configPath = filepath.Join(configDirectory, "config")
		return
	}

	// $OCI_CLI_CONFIG_FILE environment variable takes third priority
	if envConfigFile := os.Getenv("OCI_CLI_CONFIG_FILE"); envConfigFile != "" {
		configPath = envConfigFile
		configDirectory = filepath.Dir(configPath)
		return
	}

	// Default to ~/.oci/config
	home, _ := os.UserHomeDir()
	configDirectory = filepath.Join(home, ".oci")
	configPath = filepath.Join(configDirectory, "config")
	return
}

// resolveProfile determines the OCI profile from flags and environment variables.
// Priority: --profile > $OCI_CLI_PROFILE > DEFAULT
func resolveProfile() string {
	if *profile != "" {
		return *profile
	}
	if envProfile := os.Getenv("OCI_CLI_PROFILE"); envProfile != "" {
		return envProfile
	}
	return "DEFAULT"
}

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Printf("oci-tf-bootstrap %s\n", version)
		fmt.Printf("  commit: %s\n", commit)
		fmt.Printf("  built:  %s\n", date)
		os.Exit(0)
	}

	ociConfigPath, _ := resolveConfigPath()
	ociProfile := resolveProfile()

	fmt.Printf("oci-tf-bootstrap\n")
	fmt.Printf("  Profile:    %s\n", ociProfile)
	fmt.Printf("  Config:     %s\n", ociConfigPath)
	fmt.Printf("  Output:     %s\n", *outputDir)
	if *alwaysFree {
		fmt.Printf("  Mode:       always-free tier\n")
	}

	// Check if config file exists and provide helpful error message
	if _, err := os.Stat(ociConfigPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "\nError: OCI config file not found at %s\n\n", ociConfigPath)
		printSetupHelp(ociConfigPath)
		os.Exit(1)
	}

	ctx, err := discovery.NewContext(ociProfile, ociConfigPath, *region, *alwaysFree)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize OCI context: %v\n", err)
		if strings.Contains(err.Error(), "can not read") || strings.Contains(err.Error(), "configuration") {
			fmt.Fprintln(os.Stderr)
			printSetupHelp(ociConfigPath)
		}
		os.Exit(1)
	}

	fmt.Printf("  Tenancy:    %s\n", ctx.TenancyID)
	fmt.Printf("  Region:     %s\n", ctx.Region)
	fmt.Println()

	result, err := discovery.Run(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Discovery failed: %v\n", err)
		os.Exit(1)
	}

	if *jsonOut {
		if err := renderer.OutputJSON(result, os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to output JSON: %v\n", err)
			os.Exit(1)
		}
	} else {
		opts := renderer.Options{
			AlwaysFree: *alwaysFree,
		}
		if err := renderer.OutputTerraform(result, *outputDir, opts); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to render terraform: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Generated terraform files in %s\n", *outputDir)
	}
}

// printSetupHelp prints instructions for setting up OCI CLI configuration
func printSetupHelp(configPath string) {
	fmt.Fprintln(os.Stderr, "To set up OCI CLI authentication:")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "  1. Install OCI CLI:")
	fmt.Fprintln(os.Stderr, "     brew install oci-cli           # macOS")
	fmt.Fprintln(os.Stderr, "     pip install oci-cli            # pip")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "  2. Run initial setup:")
	fmt.Fprintln(os.Stderr, "     oci setup config")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "     This will prompt for:")
	fmt.Fprintln(os.Stderr, "     - Tenancy OCID (from OCI Console > Profile > Tenancy)")
	fmt.Fprintln(os.Stderr, "     - User OCID (from OCI Console > Profile > User Settings)")
	fmt.Fprintln(os.Stderr, "     - Region (e.g., us-ashburn-1)")
	fmt.Fprintln(os.Stderr, "     - API key generation")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "  3. Upload the generated public key to OCI Console:")
	fmt.Fprintln(os.Stderr, "     Profile > User Settings > API Keys > Add API Key")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Alternative: Use environment variables or flags:")
	fmt.Fprintln(os.Stderr, "  --config-file /path/to/config     # specify config file path")
	fmt.Fprintln(os.Stderr, "  --config /path/to/oci-dir         # specify config directory")
	fmt.Fprintln(os.Stderr, "  --profile PROFILE_NAME            # use specific profile")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "  OCI_CLI_CONFIG_FILE=/path/to/config")
	fmt.Fprintln(os.Stderr, "  OCI_CLI_PROFILE=PROFILE_NAME")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Documentation: https://docs.oracle.com/en-us/iaas/Content/API/Concepts/sdkconfig.htm")
}
