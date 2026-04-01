package security

import (
	"fmt"

	"github.com/pquerna/otp/totp"
	"nighttroop.t1d.nz/utils/consts"
)

func (cfg *SecurityConfig) TOTPGenerate(accountName string) (totpSecret, totpEncSecret string, err error) {
	totpKey, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "NightTroop",
		AccountName: fmt.Sprintf("%s@%s", accountName, cfg.HostName),
	})

	if err != nil {
		return "", "", fmt.Errorf("cannot generate TOTP key: %w", err)
	}

	totpSecret = totpKey.Secret()

	if cfg.Mode == consts.Development || cfg.Mode == consts.Test {
		totpSecret = consts.DEV_ADMIN_TOTP_SECRET
	}

	encTOTPSecret, err := cfg.Encrypt(totpSecret)
	if err != nil {
		return "", "", fmt.Errorf("cannot encrypt TOTP secret: %w", err)
	}

	return totpSecret, encTOTPSecret, nil
}

func (cfg *SecurityConfig) GeneratePassword() (password, hashedPassword string, err error) {
	password, hashedPassword, err = SecureCredentials()

	if cfg.Mode == consts.Development || cfg.Mode == consts.Test {
		// We are going to override this stuff in dev, just to make it easier
		password = consts.DEV_ADMIN_PASSWORD
		hashedPassword = HashPassword(password)
	}

	if err != nil {
		return "", "", fmt.Errorf("cannot generate secure credentials: %w", err)
	}
	return password, hashedPassword, nil
}

func (cfg *SecurityConfig) TOTPValidate(token, secret string) bool {
	if secret == consts.DEV_ADMIN_TOTP_SECRET {
		return token == "123123"
	}
	return totp.Validate(token, secret)
}
