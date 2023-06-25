package scheduler

type JobSetting struct {
	Disabled bool   `yaml:"disabled"`
	Schedule string `yaml:"schedule"`
	Name     string `yaml:"name"`
	Timeout  string `yaml:"timeout"`
}

type Config struct {
	Timeout     string       `yaml:"timeout"`
	JobSettings []JobSetting `yaml:"jobs"`
}
