package authenticator

import "testing"

func TestInFours(t *testing.T) {
	for _, test := range []struct{ secret, want string }{
		{"", ""},
		{"ABC", "ABC"},
		{"ABCD", "ABCD"},
		{"ABCDE", "ABCD E"},
		{"FakeTOTPSecretForDevOnly", "Fake TOTP Secr etFo rDev Only"},
		{"JBSWY3DPEHPK3PXPJBSWY3DPEHPK3PXP", "JBSW Y3DP EHPK 3PXP JBSW Y3DP EHPK 3PXP"},
	} {
		if got := inFours(test.secret); got != test.want {
			t.Errorf("inFours(%q) = %q, want %q", test.secret, got, test.want)
		}
	}
}
