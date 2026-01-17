package discovery

type Compartment struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ParentID    string `json:"parent_id"`
	Path        string `json:"path"`
}

type AvailabilityDomain struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	FaultDomains []string `json:"fault_domains"`
}

type Shape struct {
	Name           string   `json:"name"`
	ProcessorDesc  string   `json:"processor_description"`
	OCPUs          float32  `json:"ocpus"`
	MemoryGB       float32  `json:"memory_gb"`
	IsFlexible     bool     `json:"is_flexible"`
	MaxOCPUs       float32  `json:"max_ocpus,omitempty"`
	MaxMemoryGB    float32  `json:"max_memory_gb,omitempty"`
	AvailableLimit int      `json:"available_limit"`
	AvailableInADs []string `json:"available_in_ads"`
}

type Image struct {
	ID               string   `json:"id"`
	DisplayName      string   `json:"display_name"`
	OS               string   `json:"operating_system"`
	OSVersion        string   `json:"operating_system_version"`
	TimeCreated      string   `json:"time_created"`
	SizeGB           float64  `json:"size_gb"`
	CompatibleShapes []string `json:"compatible_shapes"`
}
