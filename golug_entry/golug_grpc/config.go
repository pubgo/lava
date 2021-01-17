package golug_grpc

const Name = "grpc_entry"

type Cfg struct {
	GwAddr                string `json:"gw_addr"`
	Codec                 string `json:"codec"`
	ConnectionTimeout     string `json:"connection_timeout"`
	Cp                    string `json:"cp"`
	Creds                 string `json:"creds"`
	Dc                    string `json:"dc"`
	HeaderTableSize       int64  `json:"header_table_size"`
	InitialConnWindowSize int64  `json:"initial_conn_window_size"`
	InitialWindowSize     int64  `json:"initial_window_size"`
	KeepaliveParams       struct {
		MaxConnectionAge      string `json:"max_connection_age"`
		MaxConnectionAgeGrace string `json:"max_connection_age_grace"`
		MaxConnectionIdle     string `json:"max_connection_idle"`
		Time                  string `json:"time"`
		Timeout               string `json:"timeout"`
	} `json:"keepalive_params"`
	KeepalivePolicy struct {
		MinTime             string `json:"min_time"`
		PermitWithoutStream bool   `json:"permit_without_stream"`
	} `json:"keepalive_policy"`
	MaxConcurrentStreams  int64 `json:"max_concurrent_streams"`
	MaxHeaderListSize     int64 `json:"max_header_list_size"`
	MaxReceiveMessageSize int64 `json:"max_receive_message_size"`
	MaxSendMessageSize    int64 `json:"max_send_message_size"`
	ReadBufferSize        int64 `json:"read_buffer_size"`
	WriteBufferSize       int64 `json:"write_buffer_size"`
}

const name = `
{
  "write_buffer_size": 1,
  "read_buffer_size": 1,
  "initial_window_size": 1,
  "initial_conn_window_size": 1,
  "keepalive_params": {
    "max_connection_idle": "1s",
    "max_connection_age": "2s",
    "max_connection_age_grace": "2s",
    "time": "1s",
    "timeout": "1s"
  },
  "keepalive_policy": {
    "permit_without_stream": true,
    "min_time": "1s"
  },
  "codec": "json",
  "cp": "gzip",
  "dc": "gzip",
  "max_receive_message_size": 1,
  "max_send_message_size": 1,
  "max_concurrent_streams": 1,
  "creds": "tls",
  "connection_timeout": "2s",
  "max_header_list_size": 2,
  "header_table_size": 1
}
`
