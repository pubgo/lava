package turncred_test

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"testing"
	"time"

	"github.com/pubgo/lava/v2/core/p2p/turncred"
)

func TestGenerate(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	user, pass, err := turncred.Generate("test-secret", "node-a", time.Hour, now)
	if err != nil {
		t.Fatal(err)
	}
	wantUser := "1700003600:node-a"
	if user != wantUser {
		t.Fatalf("username=%q want %q", user, wantUser)
	}

	mac := hmac.New(sha1.New, []byte("test-secret"))
	mac.Write([]byte(wantUser))
	wantPass := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	if pass != wantPass {
		t.Fatalf("password=%q want %q", pass, wantPass)
	}
}

func TestGenerateEmptySecret(t *testing.T) {
	_, _, err := turncred.Generate("", "x", time.Hour, time.Now())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGenerateDefaultUserID(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	user, _, err := turncred.Generate("s", "", time.Minute, now)
	if err != nil {
		t.Fatal(err)
	}
	if user != "1700000060:lava" {
		t.Fatalf("user=%q", user)
	}
}
