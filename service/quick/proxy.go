package quick

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"io"
	"log"
	"math/big"
	"net"
	"sync"

	 "github.com/lucas-clemente/quic-go"
)

func server(bind string) error {
	// generate TLS certificate
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return err
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return err
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return err
	}
	config := &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"quicssh"},
	}

	// configure listener
	listener, err := quic.ListenAddr(bind, config, nil)
	if err != nil {
		return err
	}
	defer listener.Close()
	log.Printf("Listening at %q...", bind)

	ctx := context.Background()
	for {
		log.Printf("Accepting connection...")
		session, err := listener.Accept(ctx)
		if err != nil {
			log.Printf("listener error: %v", err)
			continue
		}

		go serverSessionHandler(ctx, session)
	}
	return nil
}

func serverSessionHandler(ctx context.Context, session quic.Session) {
	log.Printf("hanling session...")
	for {
		stream, err := session.AcceptStream(ctx)
		if err != nil {
			log.Printf("session error: %v", err)
			break
		}
		go serverStreamHandler(ctx, stream)
	}
}

func serverStreamHandler(ctx context.Context, conn io.ReadWriteCloser) {
	log.Printf("handling stream...")
	defer conn.Close()

	rConn, err := net.DialTCP("tcp", nil, &net.TCPAddr{IP: net.IP{127, 0, 0, 1}, Port: 22})
	if err != nil {
		log.Printf("dial error: %v", err)
		return
	}
	defer rConn.Close()

	ctx, cancel := context.WithCancel(ctx)

	var wg sync.WaitGroup
	wg.Add(2)
	c1 := readAndWrite(ctx, conn, rConn, &wg)
	c2 := readAndWrite(ctx, rConn, conn, &wg)
	select {
	case err = <-c1:
		if err != nil {
			log.Printf("readAndWrite error on c1: %v", err)
			return
		}
	case err = <-c2:
		if err != nil {
			log.Printf("readAndWrite error on c2: %v", err)
			return
		}
	}
	cancel()
	wg.Wait()
	log.Printf("Piping finished")
}

func readAndWrite(ctx context.Context, r io.Reader, w io.Writer, wg *sync.WaitGroup) <-chan error {
	c := make(chan error)
	go func() {
		if wg != nil {
			defer wg.Done()
		}
		buff := make([]byte, 1024)

		for {
			select {
			case <-ctx.Done():
				return
			default:
				nr, err := r.Read(buff)
				if err != nil {
					return
				}
				if nr > 0 {
					_, err := w.Write(buff[:nr])
					if err != nil {
						return
					}
				}
			}
		}
	}()
	return c
}
