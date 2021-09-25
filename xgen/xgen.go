package xgen

type GrpcRestHandler struct {
	Service      string `json:"service"`
	Method       string `json:"method"`
	Name         string `json:"name"`
	Path         string `json:"path"`
	ClientStream bool   `json:"client_stream"`
	ServerStream bool   `json:"server_stream"`
	DefaultUrl   bool   `json:"default_url"`
}
