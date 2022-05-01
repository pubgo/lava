package env

var Cfg = struct {
	// Prefix 系统环境变量前缀
	Prefix string

	// Separator 分隔符
	Separator string
}{
	Prefix:    "",
	Separator: "_",
}
