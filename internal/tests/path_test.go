package tests

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"log"
	"net/http"
	"regexp"
	"testing"
	"time"

	nat "github.com/libp2p/go-nat"
)

func TestPath(t *testing.T) {
	re := regexp.MustCompile(`\{([^}]*)\}`)
	assert.Equal(t, re.FindStringSubmatch("123 {hello} world"), []string{"{hello}", "hello"})
	assert.Equal(t, re.FindStringSubmatch("123 {hello.abc} world"), []string{"{hello.abc}", "hello.abc"})
	assert.Equal(t, re.FindStringSubmatch("123 {hello.abc.*} world"), []string{"{hello.abc.*}", "hello.abc.*"})
	assert.Equal(t, re.FindStringSubmatch("123 {hello.abc=**} world"), []string{"{hello.abc=**}", "hello.abc=**"})
	assert.Equal(t, re.FindStringSubmatch("123 {hello.abc=*} world"), []string{"{hello.abc=*}", "hello.abc=*"})
	assert.Equal(t, re.FindAllStringSubmatch("123 {hello}/{abc} world", -1), [][]string{{"{hello}", "hello"}, {"{abc}", "abc"}})
}

func TestQuery(t *testing.T) {
	var re = regexp.MustCompile(`^(.*)\[(.*)\]$`)
	assert.Equal(t, re.FindStringSubmatch("abc[qq]"), []string{"abc[qq]", "abc", "qq"})
}

func TestName(t *testing.T) {
	nat, err := nat.DiscoverGateway(context.Background())
	if err != nil {
		log.Fatalf("error: %s", err)
	}
	log.Printf("nat type: %s", nat.Type())

	daddr, err := nat.GetDeviceAddress()
	if err != nil {
		log.Fatalf("error: %s", err)
	}
	log.Printf("device address: %s", daddr)

	iaddr, err := nat.GetInternalAddress()
	if err != nil {
		log.Fatalf("error: %s", err)
	}
	log.Printf("internal address: %s", iaddr)

	eaddr, err := nat.GetExternalAddress()
	if err != nil {
		log.Fatalf("error: %s", err)
	}
	log.Printf("external address: %s", eaddr)

	eport, err := nat.AddPortMapping(context.Background(), "tcp", 3080, "http", 60)
	if err != nil {
		log.Fatalf("error: %s", err)
	}

	log.Printf("test-page: http://%s:%d/", eaddr, eport)

	go func() {
		for {
			time.Sleep(30 * time.Second)

			_, err = nat.AddPortMapping(context.Background(), "tcp", 3080, "http", 60)
			if err != nil {
				log.Fatalf("error: %s", err)
			}
		}
	}()

	defer nat.DeletePortMapping(context.Background(), "tcp", 3080)

	http.ListenAndServe(":3080", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "text/plain")
		rw.WriteHeader(200)
		fmt.Fprintf(rw, "Hello there!\n")
		fmt.Fprintf(rw, "nat type: %s\n", nat.Type())
		fmt.Fprintf(rw, "device address: %s\n", daddr)
		fmt.Fprintf(rw, "internal address: %s\n", iaddr)
		fmt.Fprintf(rw, "external address: %s\n", eaddr)
		fmt.Fprintf(rw, "test-page: http://%s:%d/\n", eaddr, eport)
	}))
}
