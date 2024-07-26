package gateway

//var _ grpc.ServerStream = (*serverInProcess)(nil)
//
//type serverInProcess struct {
//	ctx    context.Context
//	method string
//	args   any
//	reply  any
//	opts   []grpc.CallOption
//	desc   *grpc.StreamDesc
//}
//
//func (s serverInProcess) SetHeader(md metadata.MD) error {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (s serverInProcess) SendHeader(md metadata.MD) error {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (s serverInProcess) SetTrailer(md metadata.MD) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (s serverInProcess) Context() context.Context {
//	return s.ctx
//}
//
//func (s serverInProcess) SendMsg(m any) error {
//	s.reply = m
//	return nil
//}
//
//func (s serverInProcess) RecvMsg(m any) error {
//	return inprocgrpc.ProtoCloner{}.Copy(m, s.args)
//}
