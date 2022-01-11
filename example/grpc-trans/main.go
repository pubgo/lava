package grpc_trans

// https://github.com/asim/go-micro/blob/master/plugins/server/grpc/grpc.go
//func init() {
//	grpc.UnknownServiceHandler(handler)
//}
//
//func handler(srv interface{}, stream grpc.ServerStream) error {
//	fullMethod, ok := grpc.MethodFromServerStream(stream)
//	if !ok {
//		return status.Errorf(codes.Internal, "method does not exist in context")
//	}
//
//	// get grpc metadata
//	gmd, ok := metadata.FromIncomingContext(stream.Context())
//	if !ok {
//		gmd = metadata.MD{}
//	}
//
//	// process the standard request flow
//	g.rpc.mu.Lock()
//	service := g.rpc.serviceMap[serviceName]
//	g.rpc.mu.Unlock()
//
//	if service == nil {
//		return status.New(codes.Unimplemented, fmt.Sprintf("unknown service %s", serviceName)).Err()
//	}
//
//	mtype := service.method[methodName]
//	if mtype == nil {
//		return status.New(codes.Unimplemented, fmt.Sprintf("unknown service %s.%s", serviceName, methodName)).Err()
//	}
//
//	// process unary
//	if !mtype.stream {
//		return g.processRequest(stream, service, mtype, ct, ctx)
//	}
//
//	// process stream
//	return g.processStream(stream, service, mtype, ct, ctx)
//}
//
//func processRequest(stream grpc.ServerStream, service *service, mtype *methodType, ct string, ctx context.Context) error {
//	for {
//		var argv, replyv reflect.Value
//
//		// Decode the argument value.
//		argIsValue := false // if true, need to indirect before calling.
//		if mtype.ArgType.Kind() == reflect.Ptr {
//			argv = reflect.New(mtype.ArgType.Elem())
//		} else {
//			argv = reflect.New(mtype.ArgType)
//			argIsValue = true
//		}
//
//		// Unmarshal request
//		if err := stream.RecvMsg(argv.Interface()); err != nil {
//			return err
//		}
//
//		if argIsValue {
//			argv = argv.Elem()
//		}
//
//		// reply value
//		replyv = reflect.New(mtype.ReplyType.Elem())
//
//		function := mtype.method.Func
//		var returnValues []reflect.Value
//
//		cc, err := g.newGRPCCodec(ct)
//		if err != nil {
//			return errors.InternalServerError("go.micro.server", err.Error())
//		}
//		b, err := cc.Marshal(argv.Interface())
//		if err != nil {
//			return err
//		}
//
//		// create a client.Request
//		r := &rpcRequest{
//			service:     g.opts.Name,
//			contentType: ct,
//			method:      fmt.Sprintf("%s.%s", service.name, mtype.method.Name),
//			body:        b,
//			payload:     argv.Interface(),
//		}
//
//		// define the handler func
//		fn := func(ctx context.Context, req server.Request, rsp interface{}) (err error) {
//			defer func() {
//				if r := recover(); r != nil {
//					logger.Extract(ctx).Errorf("panic recovered: %v, stack: %s", r, string(debug.Stack()))
//					err = errors.InternalServerError("go.micro.server", "panic recovered: %v", r)
//				}
//			}()
//			returnValues = function.Call([]reflect.Value{service.rcvr, mtype.prepareContext(ctx), reflect.ValueOf(argv.Interface()), reflect.ValueOf(rsp)})
//
//			// The return value for the method is an error.
//			if rerr := returnValues[0].Interface(); rerr != nil {
//				err = rerr.(error)
//			}
//
//			return err
//		}
//
//		// wrap the handler func
//		for i := len(g.opts.HdlrWrappers); i > 0; i-- {
//			fn = g.opts.HdlrWrappers[i-1](fn)
//		}
//
//		statusCode := codes.OK
//		statusDesc := ""
//		// execute the handler
//		if appErr := fn(ctx, r, replyv.Interface()); appErr != nil {
//			var errStatus *status.Status
//			switch verr := appErr.(type) {
//			case *errors.Error:
//				// micro.Error now proto based and we can attach it to grpc status
//				statusCode = microError(verr)
//				statusDesc = verr.Error()
//				verr.Detail = strings.ToValidUTF8(verr.Detail, "")
//				errStatus, err = status.New(statusCode, statusDesc).WithDetails(verr)
//				if err != nil {
//					return err
//				}
//			case proto.Message:
//				// user defined error that proto based we can attach it to grpc status
//				statusCode = convertCode(appErr)
//				statusDesc = appErr.Error()
//				errStatus, err = status.New(statusCode, statusDesc).WithDetails(verr)
//				if err != nil {
//					return err
//				}
//			default:
//				// default case user pass own error type that not proto based
//				statusCode = convertCode(verr)
//				statusDesc = verr.Error()
//				errStatus = status.New(statusCode, statusDesc)
//			}
//
//			return errStatus.Err()
//		}
//
//		if err := stream.SendMsg(replyv.Interface()); err != nil {
//			return err
//		}
//		return status.New(statusCode, statusDesc).Err()
//	}
//}
//
//func processStream(stream grpc.ServerStream, service *service, mtype *methodType, ct string, ctx context.Context) error {
//	opts := g.opts
//
//	r := &rpcRequest{
//		service:     opts.Name,
//		contentType: ct,
//		method:      fmt.Sprintf("%s.%s", service.name, mtype.method.Name),
//		stream:      true,
//	}
//
//	ss := &rpcStream{
//		request: r,
//		s:       stream,
//	}
//
//	function := mtype.method.Func
//	var returnValues []reflect.Value
//
//	// Invoke the method, providing a new value for the reply.
//	fn := func(ctx context.Context, req server.Request, stream interface{}) error {
//		returnValues = function.Call([]reflect.Value{service.rcvr, mtype.prepareContext(ctx), reflect.ValueOf(stream)})
//		if err := returnValues[0].Interface(); err != nil {
//			return err.(error)
//		}
//
//		return nil
//	}
//
//	for i := len(opts.HdlrWrappers); i > 0; i-- {
//		fn = opts.HdlrWrappers[i-1](fn)
//	}
//
//	statusCode := codes.OK
//	statusDesc := ""
//
//	if appErr := fn(ctx, r, ss); appErr != nil {
//		var err error
//		var errStatus *status.Status
//		switch verr := appErr.(type) {
//		case *errors.Error:
//			// micro.Error now proto based and we can attach it to grpc status
//			statusCode = microError(verr)
//			statusDesc = verr.Error()
//			verr.Detail = strings.ToValidUTF8(verr.Detail, "")
//			errStatus, err = status.New(statusCode, statusDesc).WithDetails(verr)
//			if err != nil {
//				return err
//			}
//		case proto.Message:
//			// user defined error that proto based we can attach it to grpc status
//			statusCode = convertCode(appErr)
//			statusDesc = appErr.Error()
//			errStatus, err = status.New(statusCode, statusDesc).WithDetails(verr)
//			if err != nil {
//				return err
//			}
//		default:
//			// default case user pass own error type that not proto based
//			statusCode = convertCode(verr)
//			statusDesc = verr.Error()
//			errStatus = status.New(statusCode, statusDesc)
//		}
//		return errStatus.Err()
//	}
//
//	return status.New(statusCode, statusDesc).Err()
//}
