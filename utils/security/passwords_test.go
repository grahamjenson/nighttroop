package security_test

import (
	"testing"
	"time"

	"github.com/pquerna/otp/totp"

	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/security"
)

// A TOTP code is six digits, so one with letters, spaces or the wrong length is refused like a wrong code
func TestTOTPValidateOnlyTakesSixDigits(t *testing.T) {
	cfg := security.NewSecurityConfig("", consts.Test, "localhost", false)
	// In tests TOTPGenerate gives the dev secret, so a real one is made here
	key, err := totp.Generate(totp.GenerateOpts{Issuer: "NightTroop", AccountName: "admin@localhost"})
	if err != nil {
		t.Fatal(err)
	}
	secret := key.Secret()
	code, err := totp.GenerateCode(secret, time.Now())
	if err != nil {
		t.Fatal(err)
	}

	devCode, err := totp.GenerateCode("FakeTOTPSecretForDevOnly", time.Now())
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name   string
		token  string
		secret string
		valid  bool
	}{
		{"the code now", code, secret, true},
		{"the code with a letter", code[:5] + "a", secret, false},
		{"letters", "abcdef", secret, false},
		{"the code with spaces", " " + code + " ", secret, false},
		{"too short", code[:5], secret, false},
		{"empty", "", secret, false},
		{"the dev code", "123123", "FakeTOTPSecretForDevOnly", true},
		{"the dev secret's code now", devCode, "FakeTOTPSecretForDevOnly", true},
		{"letters for the dev secret", "abcabc", "FakeTOTPSecretForDevOnly", false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := cfg.TOTPValidate(test.token, test.secret); got != test.valid {
				t.Errorf("TOTPValidate(%q) = %v, want %v", test.token, got, test.valid)
			}
		})
	}
}
