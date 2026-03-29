package config

import (
	"fmt"

	"github.com/pquerna/otp/totp"
	"nighttroop.t1d.nz/utils"
)

func (cfg *ServerConfig) TOTPGenerate(accountName string) (totpSecret, totpEncSecret string, err error) {
	totpKey, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "NightTroop",
		AccountName: fmt.Sprintf("%s@%s", accountName, cfg.HostName),
	})

	if err != nil {
		return "", "", fmt.Errorf("cannot generate TOTP key: %w", err)
	}

	totpSecret = totpKey.Secret()

	if cfg.Mode == Development || cfg.Mode == Test {
		totpSecret = DEV_ADMIN_TOTP_SECRET
	}

	encTOTPSecret, err := cfg.Encrypt(totpSecret)
	if err != nil {
		return "", "", fmt.Errorf("cannot encrypt TOTP secret: %w", err)
	}

	return totpSecret, encTOTPSecret, nil
}

func (cfg *ServerConfig) GeneratePassword() (password, hashedPassword string, err error) {
	password, hashedPassword, err = utils.SecureCredentials()

	if cfg.Mode == Development || cfg.Mode == Test {
		// We are going to override this stuff in dev, just to make it easier
		password = DEV_ADMIN_PASSWORD
		hashedPassword = utils.HashPassword(password)
	}

	if err != nil {
		return "", "", fmt.Errorf("cannot generate secure credentials: %w", err)
	}
	return password, hashedPassword, nil
}

func (cfg *ServerConfig) TOTPValidate(token, secret string) bool {
	if secret == DEV_ADMIN_TOTP_SECRET {
		return token == "123123"
	}
	return totp.Validate(token, secret)
}
