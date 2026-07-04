package grpccresolver

import "testing"

func TestSplit2(t *testing.T) {
	t.Parallel()
	before, after, ok := split2("a:b:c", ":")
	if !ok || before != "a" || after != "b:c" {
		t.Fatalf("split2 = %q %q %v", before, after, ok)
	}
	_, _, ok = split2("nosep", ":")
	if ok {
		t.Fatal("expected false for missing sep")
	}
}

func TestTranslateEndpointHTTP(t *testing.T) {
	t.Parallel()
	addr, name, creds := translateEndpoint("https://api.example.com:443")
	if addr != "api.example.com:443" || name != "api.example.com" || creds != CREDS_REQUIRE {
		t.Fatalf("https: addr=%q name=%q creds=%v", addr, name, creds)
	}
	addr, name, creds = translateEndpoint("http://127.0.0.1:8080")
	if addr != "127.0.0.1:8080" || name != "127.0.0.1" || creds != CREDS_DROP {
		t.Fatalf("http: addr=%q name=%q creds=%v", addr, name, creds)
	}
}

func TestInterpretPlainHostPort(t *testing.T) {
	t.Parallel()
	addr, name := Interpret("10.0.0.44:437")
	if addr != "10.0.0.44:437" || name != "10.0.0.44" {
		t.Fatalf("got addr=%q name=%q", addr, name)
	}
}

func TestSchemeToCredsRequirement(t *testing.T) {
	t.Parallel()
	if schemeToCredsRequirement("https") != CREDS_REQUIRE {
		t.Fatal("https should require creds")
	}
	if schemeToCredsRequirement("http") != CREDS_DROP {
		t.Fatal("http should drop creds")
	}
}
