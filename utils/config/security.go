package config

import (
	"fmt"

	"github.com/pquerna/otp/totp"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/config"
)

func (cfg *ServerConfig) TOTPGenerate(accountName string) (totpSecret, totpEncSecret string, err error) {
	totpKey, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "NightTroop",
		AccountName: accountName,
	})

	if err != nil {
		return "", "", fmt.Errorf("cannot generate TOTP key: %w", err)
	}

	totpSecret = totpKey.Secret()

	if cfg.Mode == config.Development {
		// We are going to override this stuff in dev, just to make it easier
		password = DEV_ADMIN_PASSWORD
		hashedPassword = utils.HashPassword(password)
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
	if cfg.Mode == config.Development {
		// We are going to override this stuff in dev, just to make it easier
		password = DEV_ADMIN_PASSWORD
		hashedPassword = utils.HashPassword(password)
	}
	if err != nil {
		return "", "", fmt.Errorf("cannot generate secure credentials: %w", err)
	}
	return password, hashedPassword, nil
}
