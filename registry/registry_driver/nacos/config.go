package nacos

const (
	Name = "nacos"
)

type Cfg struct {
	Driver      string `json:"driver"`
	NamespaceId string `json:"namespace"`
	Group       string `json:"group"`
	Cluster     string `json:"cluster"`
}
