package security

import (
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/tink-crypto/tink-go/v2/jwt"
	"github.com/tink-crypto/tink-go/v2/keyset"
	"github.com/tink-crypto/tink-go/v2/tink"
	"golang.org/x/crypto/bcrypt"

	"nighttroop.t1d.nz/utils/consts"
)

// DevKeySets are the keys test mode always uses, and .env.dev and .env.test set. They are public, so anyone could
// sign control logins, and decrypt secrets and backups, made with them: only development and test may use them
const DevKeySets = "CPrKm7wMEmQKWAowdHlwZS5nb29nbGVhcGlzLmNvbS9nb29nbGUuY3J5cHRvLnRpbmsuQWVzR2NtS2V5EiIaIEPeianjuFWmWBX0ZPsUWPIcG7VwyKUko2WSzyPDZEPNGAEQARj6ypu8DCAB,COC0mZcNEnoKbgo9dHlwZS5nb29nbGVhcGlzLmNvbS9nb29nbGUuY3J5cHRvLnRpbmsuQWVzR2NtSGtkZlN0cmVhbWluZ0tleRIrEgcIgCAQIBgDGiBAAlF6GIXYVKqNpOzDqeCHAAuQ2P2uUnlpHWdajhjzBBgBEAEY4LSZlw0gAw,CN2im5oDEmcKWwoxdHlwZS5nb29nbGVhcGlzLmNvbS9nb29nbGUuY3J5cHRvLnRpbmsuSnd0SG1hY0tleRIkEAEaIAL9oiEoyOMBd/tk4joXYbELpOdygdRSg+5u+/VbertXGAEQARjdopuaAyAB,CP7I64IGEmgKXAoudHlwZS5nb29nbGVhcGlzLmNvbS9nb29nbGUuY3J5cHRvLnRpbmsuSG1hY0tleRIoEgQIAxAgGiBaXWmgBYmmFu7Jn0GtCF1OlU5MrZuWEGdPwyzf6kjddBgBEAEY/sjrggYgAQ"

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
	// noOnesPasswordHash is checked when a login's name has no identity, so it takes as long to refuse as a wrong password
	noOnesPasswordHash string

	// totpUsedSteps is the last TOTP time step each identity signed in with, kept as long as its code is valid
	totpUsedSteps *expirable.LRU[string, uint64]
	totpMu        sync.Mutex

	// passkeySessions are the passkey ceremonies begun and not finished yet
	passkeySessions *expirable.LRU[string, []byte]

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
		passkeySessions:  newPasskeySessions(),
		totpUsedSteps:    expirable.NewLRU[string, uint64](10_000, nil, 2*time.Minute),
	}

	if cfg.IsTest() {
		// Hard code dev and tests key sets for now
		keysRaw = DevKeySets
	}

	err := cfg.ConfigCreateEncryptionKeys(keysRaw)
	if err != nil {
		panic(fmt.Sprintf("Failed to create encryption keys: %v", err))
	}

	err = cfg.createJWTPrimitives()
	if err != nil {
		panic(fmt.Sprintf("Failed to create JWT primitives: %v", err))
	}

	// A random password, at the cost of the passwords the server makes, so checking it takes as long as theirs
	cost := bcrypt.DefaultCost
	if cfg.IsDevOrTest() {
		cost = bcrypt.MinCost
	}
	_, cfg.noOnesPasswordHash, err = SecureCredentials(cost)
	if err != nil {
		panic(fmt.Sprintf("Failed to make a password hash: %v", err))
	}

	if cfg.IsDevOrTest() {
		cfg.DevelopmentDangerousHardcodedAdminPassword = new("password")
		cfg.DevelopmentDangerousHardcodedAdminTOTPSecret = new("FakeTOTPSecretForDevOnly")
		cfg.DevelopmentDangerousHardcodedAdminTOTPCode = new("123123")
	}

	return cfg
}

func (config *SecurityConfig) VerifyPassword(plainPassword, storedHash string) bool {
	return config.passwordVerifier.VerifyPassword(plainPassword, storedHash)
}

// VerifyNoOnesPassword checks plainPassword as VerifyPassword does, against a hash of no identity's password, for a
// login whose name has no identity: it is refused as slowly as a wrong password, so the time taken does not tell which
// names exist
func (config *SecurityConfig) VerifyNoOnesPassword(plainPassword string) {
	config.passwordVerifier.VerifyPassword(plainPassword, config.noOnesPasswordHash)
}

// PasswordsVerified is how many passwords VerifyPassword and VerifyNoOnesPassword have checked
func (config *SecurityConfig) PasswordsVerified() int64 {
	return config.passwordVerifier.verified.Load()
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
