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

	aeadEnc             tink.AEAD
	streamEnc           tink.StreamingAEAD
	jwtMac              jwt.MAC
	jwtControlValidator *jwt.Validator
	jwtUserValidator    *jwt.Validator
	jwtPreAuthValidator *jwt.Validator
}

func NewSecurityConfig(keysRaw string, mode consts.ServerMode, hostname string, https bool) *SecurityConfig {
	cfg := &SecurityConfig{
		Mode:     mode,
		HostName: hostname,
		HTTPS:    https,
	}

	err := cfg.ConfigCreateEncryptionKeys(keysRaw)
	if err != nil {
		panic(fmt.Sprintf("Failed to create encryption keys: %v", err))
	}

	err = cfg.createJWTPrimitives()
	if err != nil {
		panic(fmt.Sprintf("Failed to create JWT primitives: %v", err))
	}

	return cfg
}
