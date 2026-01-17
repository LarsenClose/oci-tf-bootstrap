package discovery

import (
	"fmt"
	"path/filepath"

	"github.com/oracle/oci-go-sdk/v65/common"
)

// NewContext creates a new OCI discovery context from the given config file path and profile.
// The configPath should be the full path to the OCI config file (e.g., ~/.oci/config).
func NewContext(profile, configPath, regionOverride string, alwaysFree bool) (*Context, error) {
	configProvider, err := common.ConfigurationProviderFromFileWithProfile(configPath, profile, "")
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	tenancyID, err := configProvider.TenancyOCID()
	if err != nil {
		return nil, fmt.Errorf("reading tenancy OCID from config: %w", err)
	}

	userID, err := configProvider.UserOCID()
	if err != nil {
		return nil, fmt.Errorf("reading user OCID from config: %w", err)
	}

	region, err := configProvider.Region()
	if err != nil {
		return nil, fmt.Errorf("reading region from config: %w", err)
	}

	if regionOverride != "" {
		region = regionOverride
	}

	return &Context{
		TenancyID:  tenancyID,
		UserID:     userID,
		Region:     region,
		Profile:    profile,
		ConfigPath: configPath,
		ConfigDir:  filepath.Dir(configPath),
		AlwaysFree: alwaysFree,
	}, nil
}
