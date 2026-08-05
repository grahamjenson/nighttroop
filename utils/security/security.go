package security

import (
	"fmt"

	"github.com/tink-crypto/tink-go/v2/jwt"
	"github.com/tink-crypto/tink-go/v2/keyset"
	"github.com/tink-crypto/tink-go/v2/tink"

	"nighttroop.t1d.nz/utils/consts"
)

type SecurityConfig struct {
	Mode     consts.ServerMode
	HTTPS    bool
	HostName string

	EncryptionKeySet       *keyset.Handle
	StreamEncryptionKeySet *keyset.Handle // for files
	JWTKeySet              *keyset.Handle
	MACKeySet              *keyset.Handle

	mac                 tink.MAC
	aeadEnc             tink.AEAD
	streamEnc           tink.StreamingAEAD
	jwtMac              jwt.MAC
	jwtControlValidator *jwt.Validator
	jwtUserValidator    *jwt.Validator
	jwtPreAuthValidator *jwt.Validator

	/////
	// Password Verifier
	/////

	// Still not sure this is a good idea, but legacy tokens are soooo slow
	passwordVerifier *PasswordVerifier

	/////
	// Dangerous Dev Only Configs
	/////

	DevelopmentDangerousHardcodedAdminPassword   *string
	DevelopmentDangerousHardcodedAdminTOTPSecret *string
	DevelopmentDangerousHardcodedAdminTOTPCode   *string
}

func NewSecurityConfig(keysRaw string, mode consts.ServerMode, hostname string, https bool) *SecurityConfig {
	cfg := &SecurityConfig{
		Mode:             mode,
		HostName:         hostname,
		HTTPS:            https,
		passwordVerifier: NewPasswordVerifier(),
	}

	if cfg.IsTest() {
		// Hard code dev and tests key sets for now
		keysRaw = "CPrKm7wMEmQKWAowdHlwZS5nb29nbGVhcGlzLmNvbS9nb29nbGUuY3J5cHRvLnRpbmsuQWVzR2NtS2V5EiIaIEPeianjuFWmWBX0ZPsUWPIcG7VwyKUko2WSzyPDZEPNGAEQARj6ypu8DCAB,COC0mZcNEnoKbgo9dHlwZS5nb29nbGVhcGlzLmNvbS9nb29nbGUuY3J5cHRvLnRpbmsuQWVzR2NtSGtkZlN0cmVhbWluZ0tleRIrEgcIgCAQIBgDGiBAAlF6GIXYVKqNpOzDqeCHAAuQ2P2uUnlpHWdajhjzBBgBEAEY4LSZlw0gAw,CN2im5oDEmcKWwoxdHlwZS5nb29nbGVhcGlzLmNvbS9nb29nbGUuY3J5cHRvLnRpbmsuSnd0SG1hY0tleRIkEAEaIAL9oiEoyOMBd/tk4joXYbELpOdygdRSg+5u+/VbertXGAEQARjdopuaAyAB,CP7I64IGEmgKXAoudHlwZS5nb29nbGVhcGlzLmNvbS9nb29nbGUuY3J5cHRvLnRpbmsuSG1hY0tleRIoEgQIAxAgGiBaXWmgBYmmFu7Jn0GtCF1OlU5MrZuWEGdPwyzf6kjddBgBEAEY/sjrggYgAQ"
	}

	err := cfg.ConfigCreateEncryptionKeys(keysRaw)
	if err != nil {
		panic(fmt.Sprintf("Failed to create encryption keys: %v", err))
	}

	err = cfg.createJWTPrimitives()
	if err != nil {
		panic(fmt.Sprintf("Failed to create JWT primitives: %v", err))
	}

	if cfg.IsDevOrTest() {
		cfg.DevelopmentDangerousHardcodedAdminPassword = ptr("password")
		cfg.DevelopmentDangerousHardcodedAdminTOTPSecret = ptr("FakeTOTPSecretForDevOnly")
		cfg.DevelopmentDangerousHardcodedAdminTOTPCode = ptr("123123")
	}

	return cfg
}

func (config *SecurityConfig) VerifyPassword(plainPassword, storedHash string) bool {
	return config.passwordVerifier.VerifyPassword(plainPassword, storedHash)
}

func (config *SecurityConfig) VerifyPasswordWithCache(plainPassword, storedHash string) bool {
	return config.passwordVerifier.VerifyPasswordWithCache(plainPassword, storedHash)
}

func (config *SecurityConfig) IsDev() bool {
	return config.Mode == consts.Development
}

func (config *SecurityConfig) IsProd() bool {
	return config.Mode == consts.Production
}

func (config *SecurityConfig) IsTest() bool {
	return config.Mode == consts.Test
}

func (config *SecurityConfig) IsDevOrTest() bool {
	return config.Mode == consts.Development || config.Mode == consts.Test
}

func ptr[T any](v T) *T {
	return &v
}
