package ossc

var Name = "oss"
var cfgList = make(map[string]Cfg)

type Cfg struct {
	Endpoint        string
	AccessKeyID     string
	AccessKeySecret string
	Bucket          string
}

func DefaultCfg() Cfg {
	return Cfg{}
}
