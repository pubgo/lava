package zrpc

// Route describes one unary NATS binding.
type Route struct {
	Subject string
	Queue   string
	// Register is called by generated code to bind this route on srv.
	Register func(srv *Server, queue string) error
}

// ServiceDescriptor groups routes for one protobuf service.
type ServiceDescriptor struct {
	ServicePath  string
	DefaultQueue string
	Routes       []Route
}

// RegisterService binds all routes for svc on srv.
func RegisterService(srv *Server, svc *ServiceDescriptor, queue string) error {
	if queue == "" {
		queue = svc.DefaultQueue
	}

	for _, route := range svc.Routes {
		q := route.Queue
		if q == "" {
			q = queue
		}

		if err := route.Register(srv, q); err != nil {
			return err
		}
	}

	return nil
}
