package running

import (
	"testing"

	"github.com/pubgo/funk/v2/buildinfo/version"
)

func TestGetSysInfoContainsRequiredKeys(t *testing.T) {
	info := GetSysInfo()
	required := []string{
		"project", "version", "commit_id", "instance_id",
		"grpc_port", "http_port", "debug", "hostname",
	}
	for _, key := range required {
		if _, ok := info[key]; !ok {
			t.Fatalf("GetSysInfo() missing key %q", key)
		}
	}
}

func TestCheckVersion(t *testing.T) {
	if version.Project() == "" {
		t.Skip("buildinfo not injected in test build")
	}
	defer func() {
		if recover() != nil {
			t.Fatalf("CheckVersion should not panic with valid buildinfo")
		}
	}()
	CheckVersion()
}

func TestInstanceIDNotEmpty(t *testing.T) {
	if InstanceID == "" {
		t.Fatalf("InstanceID should not be empty")
	}
}

func TestDeviceIDNotEmpty(t *testing.T) {
	if DeviceID == "" {
		t.Fatalf("DeviceID should not be empty")
	}
}
