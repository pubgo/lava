package chi

import "time"

type Cfg struct {
	Timeout   time.Duration `json:"timeout"`
	Logger    bool          `json:"logger"`
	Recover   bool          `json:"recover"`
	RequestID bool          `json:"req_id"`
}

func DefaultCfg() Cfg {
	return Cfg{
		Timeout:   60 * time.Second,
		Logger:    true,
		Recover:   true,
		RequestID: true,
	}
}
