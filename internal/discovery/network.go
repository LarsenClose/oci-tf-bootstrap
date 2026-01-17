package discovery

type VCN struct {
	ID              string           `json:"id"`
	DisplayName     string           `json:"display_name"`
	CIDRBlock       string           `json:"cidr_block"`
	CompartmentID   string           `json:"compartment_id"`
	DNSLabel        string           `json:"dns_label"`
	Subnets         []Subnet         `json:"subnets"`
	SecurityLists   []SecurityList   `json:"security_lists"`
	RouteTables     []RouteTable     `json:"route_tables"`
	InternetGateway *InternetGateway `json:"internet_gateway,omitempty"`
	NATGateway      *NATGateway      `json:"nat_gateway,omitempty"`
}

type Subnet struct {
	ID                 string `json:"id"`
	DisplayName        string `json:"display_name"`
	CIDRBlock          string `json:"cidr_block"`
	AvailabilityDomain string `json:"availability_domain"`
	IsPublic           bool   `json:"is_public"`
	DNSLabel           string `json:"dns_label"`
}

type SecurityList struct {
	ID           string         `json:"id"`
	DisplayName  string         `json:"display_name"`
	IngressRules []SecurityRule `json:"ingress_rules"`
	EgressRules  []SecurityRule `json:"egress_rules"`
}

type SecurityRule struct {
	Protocol    string `json:"protocol"`
	Source      string `json:"source,omitempty"`
	Destination string `json:"destination,omitempty"`
	PortMin     int    `json:"port_min,omitempty"`
	PortMax     int    `json:"port_max,omitempty"`
	Description string `json:"description,omitempty"`
}

type RouteTable struct {
	ID          string      `json:"id"`
	DisplayName string      `json:"display_name"`
	Routes      []RouteRule `json:"routes"`
}

type RouteRule struct {
	Destination     string `json:"destination"`
	DestinationType string `json:"destination_type"`
	NetworkEntityID string `json:"network_entity_id"`
	Description     string `json:"description,omitempty"`
}

type InternetGateway struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	IsEnabled   bool   `json:"is_enabled"`
}

type NATGateway struct {
	ID           string `json:"id"`
	DisplayName  string `json:"display_name"`
	PublicIP     string `json:"public_ip"`
	BlockTraffic bool   `json:"block_traffic"`
}

type BlockVolume struct {
	ID                 string `json:"id"`
	DisplayName        string `json:"display_name"`
	SizeGB             int64  `json:"size_gb"`
	AvailabilityDomain string `json:"availability_domain"`
	VPUsPerGB          int64  `json:"vpus_per_gb"`
	IsHydrated         bool   `json:"is_hydrated"`
}

type ServiceLimit struct {
	ServiceName  string `json:"service_name"`
	LimitName    string `json:"limit_name"`
	Value        int64  `json:"value"`
	AvailableAmt int64  `json:"available"`
	UsedAmt      int64  `json:"used"`
	Scope        string `json:"scope"`
}
