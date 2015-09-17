package models

type V3AppsModel struct {
	Name       string
	Guid       string
	Error_Code string
}

type V3PackageModel struct {
	Guid       string
	Error_Code string
}

type V3DropletModel struct {
	Guid string
}

type MetadataModel struct {
	Guid string `json:"guid"`
}

type EntityModel struct {
	Name string `json:"name"`
}
type RouteEntityModel struct {
	Host string `json:"host"`
}

type DomainsModel struct {
	NextUrl   string        `json:"next_url,omitempty"`
	Resources []DomainModel `json:"resources"`
}
type DomainModel struct {
	Metadata MetadataModel `json:"metadata"`
	Entity   EntityModel   `json:"entity"`
}
type RouteModel struct {
	Metadata MetadataModel    `json:"metadata"`
	Entity   RouteEntityModel `json:"entity"`
}
