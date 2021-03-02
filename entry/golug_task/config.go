package golug_task

const Name = "task_entry"

type Cfg struct {
	Broker    string `yaml:"broker"`
	Consumers []struct {
		Driver string `json:"driver" yaml:"driver"`
		Name   string `json:"name" yaml:"name"`
	} `json:"consumers" yaml:"consumers"`
}
