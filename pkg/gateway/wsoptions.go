package gateway

// WSConfig holds websocket frontend settings for production wiring.
type WSConfig struct {
	// OriginPatterns are passed to coder/websocket AcceptOptions.OriginPatterns.
	// When empty and InsecureSkipVerify is false, only same-origin connections are allowed.
	OriginPatterns []string
	// InsecureSkipVerify disables origin verification. Use only in development.
	InsecureSkipVerify bool
	// Subprotocols advertised during the handshake. Defaults to grpc-ws-json and grpc-ws-proto.
	Subprotocols []string
}

// WSOptionsFromConfig builds WSOption values from WSConfig.
func WSOptionsFromConfig(cfg WSConfig) []WSOption {
	opts := make([]WSOption, 0, 3)
	subprotocols := cfg.Subprotocols
	if len(subprotocols) == 0 {
		subprotocols = []string{"grpc-ws-json", "grpc-ws-proto"}
	}
	opts = append(opts, WithWSSubprotocols(subprotocols...))

	if len(cfg.OriginPatterns) > 0 {
		opts = append(opts, WithWSOriginPatterns(cfg.OriginPatterns...))
	}
	if cfg.InsecureSkipVerify {
		opts = append(opts, WithWSInsecureSkipVerify())
	}
	return opts
}
