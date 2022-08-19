package tlsutil

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"
)

func GenTLSConfig() (*tls.Config, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
	}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
	b := pem.Block{Type: "CERTIFICATE", Bytes: certDER}
	certPEM := pem.EncodeToMemory(&b)

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"quick"},
	}, nil
}
