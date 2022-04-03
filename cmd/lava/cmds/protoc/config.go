package protoc

type Cfg struct {
	Version   string   `yaml:"version,omitempty"`
	ProtoPath string   `yaml:"proto-path,omitempty"`
	Root      []string `yaml:"root,omitempty"`
	Depends   []depend `yaml:"deps,omitempty"`
	Input     []string `yaml:"input,omitempty"`
	Plugins   []plugin `yaml:"plugins,omitempty"`
}

type plugin struct {
	Name string      `yaml:"name,omitempty"`
	Out  string      `yaml:"out,omitempty"`
	Opt  interface{} `yaml:"opt,omitempty"`
}

type depend struct {
	Name    string `yaml:"name,omitempty"`
	Url     string `yaml:"url,omitempty"`
	Path    string `yaml:"path,omitempty"`
	Version string `yaml:"version,omitempty"`
}
