package golug_xgen

type GrpcRestHandler struct {
	Method        string `json:"method"`
	Name          string `json:"name"`
	Path          string `json:"path"`
	ClientStream  bool   `json:"client_stream"`
	ServerStreams bool   `json:"server_streams"`
}
