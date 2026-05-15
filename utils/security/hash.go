package security

import "encoding/base64"

func (cfg *SecurityConfig) GenerateHash(s string) (encS string, err error) {
	bytes, err := cfg.mac.ComputeMAC([]byte(s))
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil

}
