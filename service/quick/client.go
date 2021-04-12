package quick

import (
	"context"
	"crypto/tls"
	"log"
	"net"
	"os"
	"sync"

	quic "github.com/lucas-clemente/quic-go"
)

// Dial creates a new QUIC connection
// it returns once the connection is established and secured with forward-secure keys
func Dial(addr string, tlsConfig *tls.Config, quicConfig *quic.Config) (net.Conn, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}
	udpConn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		return nil, err
	}

	// DialAddr returns once a forward-secure connection is established
	quicSession, err := quic.Dial(udpConn, udpAddr, addr, tlsConfig, quicConfig)
	if err != nil {
		return nil, err
	}

	stream, err := quicSession.OpenStreamSync(context.Background())
	if err != nil {
		return nil, err
	}

	return &Conn{
		conn:    udpConn,
		session: quicSession,
		stream:  stream,
	}, nil
}


func client(addr string) error {
	ctx, cancel := context.WithCancel(context.Background())

	config := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"quicssh"},
	}

	log.Printf("Dialing %q...", addr)
	session, err := quic.DialAddr(addr, config, nil)
	if err != nil {
		return err
	}

	log.Printf("Opening stream sync...")
	stream, err := session.OpenStreamSync(ctx)
	if err != nil {
		return err
	}

	log.Printf("Piping stream with QUIC...")
	var wg sync.WaitGroup
	wg.Add(3)
	c1 := readAndWrite(ctx, stream, os.Stdout, &wg)
	c2 := readAndWrite(ctx, os.Stdin, stream, &wg)
	select {
	case err = <-c1:
		if err != nil {
			return err
		}
	case err = <-c2:
		if err != nil {
			return err
		}
	}
	cancel()
	wg.Wait()
	return nil
}
