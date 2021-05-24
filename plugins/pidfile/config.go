package pidfile

const Name = "pidfile"

type Cfg struct {
	PidPath string `json:"pidPath"`
}
