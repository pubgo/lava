package nacos

const Name = "nacos"

type Cfg struct {
	Driver      string `json:"driver"`
	Group       string `json:"group"`
	NamespaceId string `json:"namespace"`
}
