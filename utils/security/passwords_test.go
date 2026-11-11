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

// A TOTP code signs an identity in once, so one seen as it was typed cannot sign in again while it is valid, nor can
// an earlier one, though another identity with the same secret signs in with it. The fixed dev code is not one time
func TestTOTPSignInRefusesACodeUsedAlready(t *testing.T) {
	cfg := security.NewSecurityConfig("", consts.Test, "localhost", false)
	secret := "JBSWY3DPEHPK3PXP"
	now := time.Now()
	code, err := totp.GenerateCode(secret, now)
	if err != nil {
		t.Fatal(err)
	}
	earlier, err := totp.GenerateCode(secret, now.Add(-30*time.Second))
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		identity string
		token    string
		secret   string
		valid    bool
	}{
		{"the code now", "site abcdefgh 1", code, secret, true},
		{"the code again", "site abcdefgh 1", code, secret, false},
		{"the code before it", "site abcdefgh 1", earlier, secret, false},
		{"the code for another identity", "site abcdefgh 2", code, secret, true},
		{"the code for a control identity", "control 1", code, secret, true},
		{"a wrong code", "site abcdefgh 3", code[:5] + "x", secret, false},
		{"the dev code", "site abcdefgh 1", "123123", "FakeTOTPSecretForDevOnly", true},
		{"the dev code again", "site abcdefgh 1", "123123", "FakeTOTPSecretForDevOnly", true},
	}
	for _, test := range tests {
		if got := cfg.TOTPSignIn(test.identity, test.token, test.secret); got != test.valid {
			t.Errorf("%s: TOTPSignIn(%q, %q) = %v, want %v", test.name, test.identity, test.token, got, test.valid)
		}
	}
	// Checking a code, as setup does, does not use it up
	if !cfg.TOTPValidate(code, secret) {
		t.Errorf("TOTPValidate(%q) refused a code signed in with", code)
	}
}
