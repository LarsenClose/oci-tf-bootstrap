# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- `--version` flag to display version, commit, and build date information
- Block volume discovery with size and availability domain information
- Security list discovery with ingress/egress rule counts
- Route table discovery with route information
- Internet gateway discovery with enabled/disabled status
- NAT gateway discovery with public IP addresses
- Compartment hierarchy rendering (shows parent-child relationships in comments)
- Shell completions for bash, zsh, and fish
- Comprehensive test coverage for renderer components (network.go, data.go, templates.go)

### Changed
- Compartments are now displayed hierarchically in locals.tf output
- Image discovery now properly paginates through all results
- Improved error handling in context initialization (no longer silently ignores errors)

### Fixed
- Go version mismatch between CI (1.24) and release (1.22) workflows
- Image pagination - previously only processed first page of results
- Error handling in context.go now properly returns errors instead of silently ignoring them

## [0.1.0] - Initial Release

### Added
- Core OCI resource discovery (tenancy, compartments, ADs, shapes, images, VCNs, subnets)
- Terraform file generation (provider.tf, locals.tf, data.tf, instance_example.tf, network.tf)
- Always-free tier mode with `--always-free` flag
- JSON output mode with `--json` flag
- Profile and region override options
- Parallel discovery using errgroup for performance
- Cross-platform builds (darwin/linux, amd64/arm64)
- GitHub Actions CI/CD pipeline
- GoReleaser integration for releases
