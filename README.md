# oci-tf-bootstrap

[![CI](https://github.com/larsenclose/oci-tf-bootstrap/actions/workflows/ci.yml/badge.svg)](https://github.com/larsenclose/oci-tf-bootstrap/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/larsenclose/oci-tf-bootstrap)](https://goreportcard.com/report/github.com/larsenclose/oci-tf-bootstrap)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)

Solve the OCI "Blank Page Problem" - instantly generate ready-to-use Terraform configurations from your OCI tenancy.

## The Problem

Unlike AWS where `ami-ubuntu-latest` often works, OCI requires:
- Specific OCIDs that vary by region
- Tenancy-specific Availability Domain names (e.g., `GqIf:US-ASHBURN-AD-1`)
- Shape validation against service limits
- Manual hunting through console for compartment IDs

## The Solution

Run one command, get immediately usable Terraform:

```bash
oci-tf-bootstrap --output ./terraform
```

Generates:
- `provider.tf` - Configured provider block
- `locals.tf` - All discovered OCIDs as local values
- `data.tf` - Dynamic data sources that stay valid as images update
- `instance_example.tf` - Ready-to-deploy example instance

## Installation

### Homebrew (macOS/Linux)

```bash
brew install larsenclose/tap/oci-tf-bootstrap
```

### From Source

```bash
go install github.com/larsenclose/oci-tf-bootstrap@latest
```

### From Release

Download the latest binary from [Releases](https://github.com/larsenclose/oci-tf-bootstrap/releases).

### Build Locally

```bash
git clone https://github.com/larsenclose/oci-tf-bootstrap.git
cd oci-tf-bootstrap
make build
make install  # copies to ~/bin
```

## Usage

```bash
# Default: uses ~/.oci/config [DEFAULT] profile
oci-tf-bootstrap

# Specify profile and output directory
oci-tf-bootstrap --profile PROD --output ./infra/oci

# Generate only always-free tier eligible resources
oci-tf-bootstrap --always-free --output ./terraform

# JSON output for scripting
oci-tf-bootstrap --json > discovery.json
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--profile` | `DEFAULT` | OCI config profile name |
| `--config` | `~/.oci` | OCI config directory |
| `--config-file` | `~/.oci/config` | OCI config file path (takes precedence over `--config`) |
| `--output` | `./terraform` | Output directory for generated TF files |
| `--region` | from config | Override region |
| `--always-free` | `false` | Filter to always-free tier resources only |
| `--json` | `false` | Output raw discovery as JSON |

### Environment Variables

Supports standard OCI CLI environment variables:

| Variable | Description |
|----------|-------------|
| `OCI_CLI_CONFIG_FILE` | Path to OCI config file (same as `--config-file`) |
| `OCI_CLI_PROFILE` | Profile name to use (same as `--profile`) |

**Priority order:** CLI flags > environment variables > defaults

### Multiple Configurations

For users managing multiple OCI tenancies or environments:

```bash
# Using different config files
oci-tf-bootstrap --config-file ~/.oci/prod-config --output ./terraform-prod
oci-tf-bootstrap --config-file ~/.oci/dev-config --output ./terraform-dev

# Using profiles in the same config file
oci-tf-bootstrap --profile PROD --output ./terraform-prod
oci-tf-bootstrap --profile DEV --output ./terraform-dev

# Using environment variables (useful for CI/CD)
export OCI_CLI_CONFIG_FILE=/secrets/oci/config
export OCI_CLI_PROFILE=PROD
oci-tf-bootstrap --output ./terraform

# Different config directories entirely
oci-tf-bootstrap --config ~/work/.oci --profile CLIENT_A
oci-tf-bootstrap --config ~/personal/.oci --profile DEFAULT
```

**Tip:** OCI config files support multiple profiles. A common pattern:

```ini
# ~/.oci/config
[DEFAULT]
user=ocid1.user.oc1..personal
tenancy=ocid1.tenancy.oc1..personal
region=us-ashburn-1
key_file=~/.oci/personal_api_key.pem
fingerprint=aa:bb:cc:...

[WORK]
user=ocid1.user.oc1..work
tenancy=ocid1.tenancy.oc1..work
region=us-phoenix-1
key_file=~/.oci/work_api_key.pem
fingerprint=dd:ee:ff:...

[CLIENT_PROJECT]
user=ocid1.user.oc1..client
tenancy=ocid1.tenancy.oc1..client
region=eu-frankfurt-1
key_file=~/.oci/client_api_key.pem
fingerprint=11:22:33:...
```

## Always-Free Tier Mode

The `--always-free` flag filters output to only OCI always-free eligible resources:

```bash
oci-tf-bootstrap --always-free --output ./terraform
```

### What's Included

**Compute:**
- `VM.Standard.A1.Flex` (ARM): 4 OCPUs + 24GB memory total
- `VM.Standard.E2.1.Micro` (x86): 2 instances total

**Storage:**
- Block Storage: 200GB total (boot + block volumes)
- Object Storage: 20GB (after trial expires)

**Networking (all free):**
- 2 VCNs with internet/NAT/service gateways
- 1 Flexible Load Balancer (10 Mbps)
- 1 Network Load Balancer
- Bastion service
- Site-to-Site VPN (up to 50 IPSec connections)

**Generated Example:**
```hcl
resource "oci_core_instance" "always_free" {
  shape = "VM.Standard.A1.Flex"

  shape_config {
    ocpus         = 2   # Half of 4 free (allows 2 instances)
    memory_in_gbs = 12  # Half of 24GB free
  }

  source_details {
    source_id               = data.oci_core_images.canonical_ubuntu_24_04_minimal_aarch64.images[0].id
    boot_volume_size_in_gbs = 50  # Counts toward 200GB limit
  }
}
```

## Generated Output Example

### locals.tf
```hcl
locals {
  tenancy_ocid = "ocid1.tenancy.oc1..aaaa..."

  # Compartments
  comp_network    = "ocid1.compartment.oc1..bbbb..."
  comp_production = "ocid1.compartment.oc1..cccc..."

  # Availability Domains (tenancy-specific)
  ad_1 = "GqIf:US-ASHBURN-AD-1"
  ad_2 = "GqIf:US-ASHBURN-AD-2"

  # Validated Shapes
  shape_vm_standard_a1_flex = "VM.Standard.A1.Flex"  # Flex: 1-80 OCPU
}
```

### data.tf
```hcl
data "oci_core_images" "canonical_ubuntu_22_04" {
  compartment_id           = local.tenancy_ocid
  operating_system         = "Canonical Ubuntu"
  operating_system_version = "22.04"
  sort_by                  = "TIMECREATED"
  sort_order               = "DESC"
  state                    = "AVAILABLE"
}

output "latest_images" {
  value = {
    canonical_ubuntu_22_04 = data.oci_core_images.canonical_ubuntu_22_04.images[0].id
  }
}
```

## Prerequisites

### OCI CLI Setup

If you haven't configured OCI CLI yet, follow these steps:

1. **Install OCI CLI:**

   ```bash
   # macOS (Homebrew)
   brew install oci-cli

   # Linux/macOS (pip)
   pip install oci-cli

   # Windows
   # Download installer from Oracle
   ```

2. **Run initial configuration:**

   ```bash
   oci setup config
   ```

   This prompts for:
   - **Tenancy OCID**: OCI Console > Profile (top-right) > Tenancy: \<name\>
   - **User OCID**: OCI Console > Profile > User Settings
   - **Region**: e.g., `us-ashburn-1`, `us-phoenix-1`, `eu-frankfurt-1`
   - **API key**: Accept the default to generate a new key pair

3. **Upload the public key to OCI Console:**

   ```bash
   cat ~/.oci/oci_api_key_public.pem
   ```

   Copy this key, then: OCI Console > Profile > User Settings > API Keys > Add API Key > Paste Public Key

4. **Verify configuration:**

   ```bash
   oci iam region list --output table
   ```

### IAM Policy

Ensure your user has read access to resources:

```hcl
Allow group <your-group> to read all-resources in tenancy
```

### Requirements Summary

- OCI CLI configured (`~/.oci/config`) - see setup above
- IAM policy with read access to tenancy resources
- Go 1.24+ (only for building from source)

**Note:** Generated Terraform is compatible with both Terraform and OpenTofu.

## Shell Completions

Shell completion scripts are included for bash, zsh, and fish.

### Bash

```bash
# Add to ~/.bashrc or ~/.bash_profile
source /path/to/oci-tf-bootstrap/completions/oci-tf-bootstrap.bash

# Or copy to system completions directory
sudo cp completions/oci-tf-bootstrap.bash /etc/bash_completion.d/oci-tf-bootstrap
```

### Zsh

```zsh
# Add to ~/.zshrc (before compinit)
fpath=(/path/to/oci-tf-bootstrap/completions $fpath)
autoload -Uz compinit && compinit

# Or copy to site-functions
sudo cp completions/oci-tf-bootstrap.zsh /usr/local/share/zsh/site-functions/_oci-tf-bootstrap
```

### Fish

```fish
# Copy to fish completions directory
cp completions/oci-tf-bootstrap.fish ~/.config/fish/completions/
```

## Design Decisions

1. **Data Sources over hardcoded OCIDs**: Image OCIDs rot. Data sources stay valid.
2. **Parallel discovery**: Uses errgroup for concurrent API calls
3. **Service limit validation**: Only shows shapes you can actually use
4. **Single binary**: Cross-compile for Mac/Linux/ARM with no runtime deps
5. **Always-free awareness**: First-class support for cost-conscious deployments

## Cross-Compilation

```bash
make darwin   # macOS (arm64 + amd64)
make linux    # Linux x86_64
make pi       # Linux ARM64 (Raspberry Pi, ARM servers)
make all      # All platforms
```

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

Apache 2.0 - see [LICENSE](LICENSE) for details.
