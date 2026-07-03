package running

import "testing"

func TestEnvHelpers(t *testing.T) {
	orig := Env.String()
	t.Cleanup(func() { _ = Env.Set(orig) })

	cases := []struct {
		env           string
		dev           bool
		test          bool
		stage         bool
		prod          bool
		nonProd       bool
	}{
		{EnvDev, true, false, false, false, true},
		{EnvTest, false, true, false, false, true},
		{EnvStage, false, false, true, false, true},
		{EnvProd, false, false, false, true, false},
		{"custom", false, false, false, false, false},
	}

	for _, tc := range cases {
		if err := Env.Set(tc.env); err != nil {
			t.Fatalf("Env.Set(%q): %v", tc.env, err)
		}
		if got := IsDev(); got != tc.dev {
			t.Fatalf("IsDev()@%s = %v, want %v", tc.env, got, tc.dev)
		}
		if got := IsTest(); got != tc.test {
			t.Fatalf("IsTest()@%s = %v, want %v", tc.env, got, tc.test)
		}
		if got := IsStage(); got != tc.stage {
			t.Fatalf("IsStage()@%s = %v, want %v", tc.env, got, tc.stage)
		}
		if got := IsProd(); got != tc.prod {
			t.Fatalf("IsProd()@%s = %v, want %v", tc.env, got, tc.prod)
		}
		if got := IsNonProd(); got != tc.nonProd {
			t.Fatalf("IsNonProd()@%s = %v, want %v", tc.env, got, tc.nonProd)
		}
	}
}
