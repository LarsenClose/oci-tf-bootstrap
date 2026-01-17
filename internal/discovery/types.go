package discovery

type Context struct {
	TenancyID  string
	UserID     string
	Region     string
	Profile    string
	ConfigPath string // Full path to config file (e.g., ~/.oci/config)
	ConfigDir  string // Directory containing config file (e.g., ~/.oci)
	AlwaysFree bool
}

type Result struct {
	Tenancy             TenancyInfo          `json:"tenancy"`
	Compartments        []Compartment        `json:"compartments"`
	AvailabilityDomains []AvailabilityDomain `json:"availability_domains"`
	Shapes              []Shape              `json:"shapes"`
	Images              []Image              `json:"images"`
	VCNs                []VCN                `json:"vcns"`
	BlockVolumes        []BlockVolume        `json:"block_volumes"`
	Limits              []ServiceLimit       `json:"limits"`
}

type TenancyInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	HomeRegion  string `json:"home_region"`
	Description string `json:"description"`
}
