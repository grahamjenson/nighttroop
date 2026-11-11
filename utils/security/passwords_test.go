package security_test

import (
	"testing"
	"time"

	"github.com/pquerna/otp"
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

// A QR code's URI is one an authenticator app adds an account from: an otpauth://totp key with the secret, issued by
// Nighttroop, for the name at the site's host, whose codes are the secret's
func TestTOTPURIIsAKeyAnAppAdds(t *testing.T) {
	secret := "JBSWY3DPEHPK3PXPJBSWY3DPEHPK3PXP"
	for _, test := range []struct{ name, siteURL, account string }{
		{"admin", "https://abcdefgh.nighttroop.nz", "admin@abcdefgh.nighttroop.nz"},
		{"admin", "http://control.localhost:3000/setup", "admin@control.localhost:3000"},
		{"a carer", "https://abcdefgh.nighttroop.nz", "a carer@abcdefgh.nighttroop.nz"},
		{"admin", "", "admin"},
	} {
		key, err := otp.NewKeyFromURL(security.TOTPURI(test.name, test.siteURL, secret))
		if err != nil {
			t.Fatalf("%s at %q: %v", test.name, test.siteURL, err)
		}
		if key.Type() != "totp" || key.Secret() != secret || key.Issuer() != security.TOTPIssuer || key.AccountName() != test.account {
			t.Errorf("%s at %q: got %s %s %s %q, want totp %s %s %q", test.name, test.siteURL, key.Type(), key.Secret(), key.Issuer(), key.AccountName(), secret, security.TOTPIssuer, test.account)
		}
		code, err := totp.GenerateCode(key.Secret(), time.Now())
		if err != nil || !totp.Validate(code, secret) {
			t.Errorf("%s at %q: the key's code %s is not the secret's (%v)", test.name, test.siteURL, code, err)
		}
	}
}
