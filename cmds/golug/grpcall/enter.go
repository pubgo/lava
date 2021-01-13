package grpcall

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/jhump/protoreflect/grpcreflect"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	reflectpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
)

type EngineHandler struct {
	// grpc clients
	clients     map[string]*grpc.ClientConn
	clientsLock sync.RWMutex

	eventHandler InvocationEventHandler
	invokeCtl    *InvokeHandler

	dialTime      time.Duration
	keepAliveTime time.Duration
	typeCacher    *protoTypesCache

	ctx        context.Context
	cancel     context.CancelFunc
	descSource DescriptorSource
}

type Option func(*EngineHandler) error

func New(options ...Option) (*EngineHandler, error) {
	e := new(EngineHandler)

	// default values
	e.ctx, e.cancel = context.WithCancel(context.Background())
	e.dialTime = 10 * time.Second
	e.keepAliveTime = 64 * time.Second
	e.eventHandler = defaultInEventHooker
	e.clients = make(map[string]*grpc.ClientConn, 10)
	e.typeCacher = newProtoTypeCache()

	for _, opt := range options {
		if opt != nil {
			if err := opt(e); err != nil {
				return nil, err
			}
		}
	}

	return e, nil
}

func (e *EngineHandler) DoConnect(target string) (*grpc.ClientConn, error) {
	e.clientsLock.RLock() // read lock
	if conn, ok := e.clients[target]; ok {
		e.clientsLock.RUnlock()
		return conn, nil
	}

	e.clientsLock.RUnlock()

	ctx, _ := context.WithTimeout(e.ctx, e.dialTime)

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithKeepaliveParams(keepalive.ClientParameters{
		Time:    e.keepAliveTime,
		Timeout: e.keepAliveTime,
	}))

	cc, err := BlockingDial(ctx, target, opts...)
	if err != nil {
		return cc, err
	}

	e.clientsLock.Lock() // write lock
	defer e.clientsLock.Unlock()

	e.clients[target] = cc
	return cc, err
}

func (e *EngineHandler) Init() error {
	return nil
}

func (e *EngineHandler) Close() error {
	e.cancel()

	e.clientsLock.Lock()
	defer e.clientsLock.Unlock()
	for _, cc := range e.clients {
		cc.Close()
	}

	return nil
}

func (e *EngineHandler) CallWithCtx(ctx context.Context, target, serviceName, methodName, data string) (*ResultModel, error) {
	return e.invokeCall(ctx, nil, target, serviceName, methodName, data)
}

func (e *EngineHandler) Call(target, serviceName, methodName, data string) (*ResultModel, error) {
	return e.invokeCall(e.ctx, nil, target, serviceName, methodName, data)
}

func (e *EngineHandler) CallWithAddr(target, serviceName, methodName, data string) (*ResultModel, error) {
	return e.invokeCall(e.ctx, nil, target, serviceName, methodName, data)
}

func (e *EngineHandler) CallWithAddrCtx(ctx context.Context, target, serviceName, methodName, data string) (*ResultModel, error) {
	return e.invokeCall(ctx, nil, target, serviceName, methodName, data)
}

func (e *EngineHandler) CallWithClient(client *grpc.ClientConn, serviceName, methodName, data string) (*ResultModel, error) {
	return e.CallWithClientCtx(nil, client, serviceName, methodName, data)
}

func (e *EngineHandler) CallWithClientCtx(ctx context.Context, client *grpc.ClientConn, serviceName, methodName, data string) (*ResultModel, error) {
	if client == nil {
		return nil, errors.New("invalid grpc client")
	}

	return e.invokeCall(ctx, client, "", serviceName, methodName, data)
}

// invokeCall request target grpc server
func (e *EngineHandler) invokeCall(ctx context.Context, gclient *grpc.ClientConn, target, serviceName, methodName, data string) (*ResultModel, error) {
	if serviceName == "" || methodName == "" {
		return nil, errors.New("serverName or methodName is null")
	}

	if gclient == nil && target == "" {
		return nil, errors.New("target addr is null")
	}

	if ctx == nil {
		ctx = e.ctx
	}

	var (
		err       error
		cc        *grpc.ClientConn
		connErr   error
		refClient *grpcreflect.Client

		addlHeaders multiString
		rpcHeaders  multiString
		reflHeaders multiString

		descSource DescriptorSource
	)

	// parse proto by grpc reflet api
	md := MetadataFromHeaders(append(addlHeaders, reflHeaders...))
	refCtx := metadata.NewOutgoingContext(e.ctx, md)
	cc, connErr = e.DoConnect(target)
	if connErr != nil {
		return nil, connErr
	}
	refClient = grpcreflect.NewClient(refCtx, reflectpb.NewServerReflectionClient(cc))
	descSource = DescriptorSourceFromServer(e.ctx, refClient)

	if gclient == nil {
		cc, connErr = e.DoConnect(target)
		if connErr != nil {
			return nil, connErr
		}
	} else {
		cc = gclient
	}

	var inData io.Reader
	inData = strings.NewReader(data)
	rf, err := RequestParserFor(descSource, inData)
	if err != nil {
		return nil, errors.New("request parse and format failed")
	}

	result, err := e.invokeCtl.InvokeRPC(e.ctx, descSource, cc, serviceName, methodName,
		append(addlHeaders, rpcHeaders...),
		rf.Next,
	)
	return result, err
}

func (e *EngineHandler) ListServices() ([]string, error) {
	return e.descSource.ListServices()
}

func (e *EngineHandler) ListMethods(svc string) ([]string, error) {
	return ListMethods(e.descSource, svc)
}

type ServMethodModel struct {
	PackageName     string
	ServiceName     string
	FullServiceName string
	MethodName      string
	FullMethodName  string
}

func (e *EngineHandler) ListServiceAndMethods() (map[string][]ServMethodModel, error) {
	servList, err := e.ListServices()
	if err != nil {
		return nil, err
	}

	m := map[string][]ServMethodModel{}
	for _, svc := range servList {
		fullMethodList, err := e.ListMethods(svc)
		servMethodModelList := []ServMethodModel{}
		for _, method := range fullMethodList {
			cs := strings.Split(method, ".")
			if len(cs) < 3 {
				return nil, errors.New("method split failed")
			}

			dto := ServMethodModel{
				MethodName:      cs[len(cs)-1],
				ServiceName:     cs[len(cs)-2],
				PackageName:     strings.Join(cs[:len(cs)-2], "."),
				FullMethodName:  method,
				FullServiceName: svc,
			}
			servMethodModelList = append(servMethodModelList, dto)
		}

		if err != nil {
			return nil, err
		}

		m[svc] = servMethodModelList
	}

	return m, nil
}

func (e *EngineHandler) ExtractProtoType(svc, mth string) (proto.Message, proto.Message, error) {
	// get types from cache
	key := e.typeCacher.makeKey(svc, mth)
	model, ok := e.typeCacher.get(key)
	if ok {
		return model.reqType, model.respType, nil
	}

	dsc, err := e.descSource.FindSymbol(svc)
	if err != nil {
		if isNotFoundError(err) {
			return nil, nil, errors.New("not find service in pb descriptor")
		}

		return nil, nil, errors.New("query service failed in pb descriptor")
	}

	sd, ok := dsc.(*desc.ServiceDescriptor)
	if !ok {
		return nil, nil, errors.New("not expose service")
	}

	mtd := sd.FindMethodByName(mth)
	if mtd == nil {
		return nil, nil, fmt.Errorf("service %q does not include a method named %q", svc, mth)
	}

	var ext dynamic.ExtensionRegistry
	alreadyFetched := map[string]bool{}
	if err = fetchAllExtensions(e.descSource, &ext, mtd.GetInputType(), alreadyFetched); err != nil {
		return nil, nil, fmt.Errorf("error resolving server extensions for message %s: %v", mtd.GetInputType().GetFullyQualifiedName(), err)
	}

	if err = fetchAllExtensions(e.descSource, &ext, mtd.GetOutputType(), alreadyFetched); err != nil {
		return nil, nil, fmt.Errorf("error resolving server extensions for message %s: %v", mtd.GetOutputType().GetFullyQualifiedName(), err)
	}

	msgFactory := dynamic.NewMessageFactoryWithExtensionRegistry(&ext)
	req := msgFactory.NewMessage(mtd.GetInputType())
	reply := msgFactory.NewMessage(mtd.GetOutputType())

	// set types to cache
	e.typeCacher.set(key, req, reply)
	return req, reply, nil
}

type multiString []string

func (s *multiString) String() string {
	return strings.Join(*s, ",")
}

func (s *multiString) IsEmpty() bool {
	if len(*s) > 0 {
		return false
	}

	return true
}

func (s *multiString) Set(value string) error {
	*s = append(*s, value)
	return nil
}
