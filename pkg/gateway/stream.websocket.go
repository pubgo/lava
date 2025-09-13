package gateway

//const (
//	// Time allowed to write a message to the peer.
//	writeWait = 10 * time.Second
//
//	// Time allowed to read the next pong message from the peer.
//	pongWait = 60 * time.Second
//
//	// Send pings to peer with this period. Must be less than pongWait.
//	pingPeriod = (pongWait * 9) / 10
//
//	// Maximum message size allowed from peer.
//	maxMessageSize = 1024 * 10
//)
//
//type streamWS struct {
//	handler        *fiber.Ctx
//	conn       *websocket.Conn
//	pathRule   *httpPathRule
//	params     url.Values
//	sentHeader bool
//	header     metadata.MD
//	trailer    metadata.MD
//}
//
//func (s *streamWS) SetHeader(md metadata.MD) error {
//	if !s.sentHeader {
//		s.header = metadata.Join(s.header, md)
//	}
//	return nil
//}
//
//func (s *streamWS) SendHeader(md metadata.MD) error {
//	if s.sentHeader {
//		return nil // already sent?
//	}
//	// TODO: headers?
//	s.sentHeader = true
//	return nil
//}
//
//func (s *streamWS) SetTrailer(md metadata.MD) {
//	s.sentHeader = true
//	s.trailer = metadata.Join(s.trailer, md)
//}
//
//func (s *streamWS) Context() context.Context {
//	sts := &serverTransportStream{ServerStream: s, method: s.pathRule.GrpcMethodName}
//	return grpc.NewContextWithServerTransportStream(s.handler.Context(), sts)
//}
//
//func (s *streamWS) SendMsg(v interface{}) error {
//	reply := v.(proto.Message)
//
//	cur := reply.ProtoReflect()
//	for _, fd := range s.pathRule.rspBody {
//		cur = cur.Mutable(fd).Message()
//	}
//	msg := cur.Interface()
//
//	// TODO: contentType check?
//	b, err := protojson.Marshal(msg)
//	if err != nil {
//		return err
//	}
//	return s.conn.WriteMessage(websocket.TextMessage, b)
//}
//
//func (s *streamWS) RecvMsg(m interface{}) error {
//	args := m.(proto.Message)
//
//	if s.pathRule.HasReqBody {
//		cur := args.ProtoReflect()
//		for _, fd := range s.pathRule.reqBody {
//			cur = cur.Mutable(fd).Message()
//		}
//
//		msg := cur.Interface()
//		_, message, err := s.conn.ReadMessage()
//		if err != nil {
//			return err
//		}
//
//		if err := protojson.Unmarshal(message, msg); err != nil {
//			return errors.Wrap(err, "failed to unmarshal protobuf json message")
//		}
//	}
//
//	if len(s.params) > 0 {
//		if err := PopulateQueryParameters(args, s.params, utilities.NewDoubleArray(nil)); err != nil {
//			log.Err(err).Msg("failed to set params")
//		}
//	}
//
//	return nil
//}
