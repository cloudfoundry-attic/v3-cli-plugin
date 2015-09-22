package models

type V3AppModel struct {
	Name       string
	Guid       string
	Error_Code string
	Processes  string
	Instances  int `json:"total_desired_instances"`
}

type V3ProcessModel struct {
	Type      string
	Instances int
	Memory    int        `json:"memory_in_mb"`
	Disk      int        `json:"disk_in_mb"`
	Links     LinksModel `json:"_links"`
}

type V3AppsModel struct {
	Apps []V3AppModel `json:"resources"`
}

type V3ProcessesModel struct {
	Processes []V3ProcessModel `json:"resources"`
}

type LinkModel struct {
	Href string
}

type LinksModel struct {
	App   LinkModel
	Space LinkModel
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
