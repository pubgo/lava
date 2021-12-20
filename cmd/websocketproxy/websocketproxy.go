package websocketproxy

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"nhooyr.io/websocket"
)

type WebSocketProxyServer struct {
	stopChan       chan struct{}
	errorChan      chan error
	connectionChan chan net.Conn
	srv            *http.Server
	laddr          string
}

func NewWebSocketProxyServer(laddr string) *WebSocketProxyServer {
	return &WebSocketProxyServer{
		laddr:          laddr,
		stopChan:       make(chan struct{}),
		errorChan:      make(chan error, 1),
		connectionChan: make(chan net.Conn)}
}

func (p *WebSocketProxyServer) Listen() (net.Listener, error) {
	p.srv = &http.Server{
		Handler: p,
	}

	lis, err := net.Listen("tcp", p.laddr)
	if err != nil {
		return nil, err
	}
	go func() {
		defer close(p.errorChan)
		p.errorChan <- p.srv.Serve(lis)
	}()

	return p, nil
}

func (p *WebSocketProxyServer) ServeHTTP(wr http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(wr, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})

	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close(websocket.StatusInternalError, "fail")

	ctx := r.Context()
	select {
	case <-p.stopChan:
		return
	default:
		p.connectionChan <- websocket.NetConn(ctx, conn, websocket.MessageBinary)
		select {
		case <-p.stopChan:
		case <-r.Context().Done():
		}
	}
	conn.Close(websocket.StatusNormalClosure, "ok")
}

func (p *WebSocketProxyServer) Accept() (net.Conn, error) {
	select {
	case <-p.stopChan:
		return nil, fmt.Errorf("server stopped")
	case err := <-p.errorChan:
		_ = p.Close()
		return nil, err
	case c := <-p.connectionChan:
		return c, nil
	}
}

func (p *WebSocketProxyServer) Close() error {
	select {
	case <-p.stopChan:
	default:
		close(p.stopChan)
	}
	if p.srv != nil {
		return p.srv.Close()
	}

	return nil
}

func (p *WebSocketProxyServer) Addr() net.Addr {
	return net.Addr(nil)
}
