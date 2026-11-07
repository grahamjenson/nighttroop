package security

import (
	"fmt"
	"sync"
	"time"

	"github.com/tink-crypto/tink-go/v2/jwt"
)

// AssumeAudience is a link a control admin follows to sign in to a site as one of its identities
var AssumeAudience = "assume"

// assumeFor is how long a link to sign in as a site's identity is good for, long enough to open it from control
const assumeFor = time.Minute

// assumeUsed are the IDs of the links followed, kept until they expire, so each signs in only once
var assumeUsed = struct {
	sync.Mutex
	until map[string]time.Time
}{until: map[string]time.Time{}}

// GenerateAssumeJWTToken is a link token for the control admin by to sign in to the site as its identity, good once
// and for a minute
func (config *SecurityConfig) GenerateAssumeJWTToken(subdomain string, identityName string, by string) (string, error) {
	now := time.Now()
	expiresAt, id := now.Add(assumeFor), SecureRandomString(32, API_KEY_CHAR_SET)
	rawJWT, err := jwt.NewRawJWT(&jwt.RawJWTOptions{Subject: &identityName, Issuer: &subdomain, Audience: &AssumeAudience,
		ExpiresAt: &expiresAt, IssuedAt: &now, JWTID: &id, CustomClaims: map[string]any{"by": by}})
	if err != nil {
		return "", err
	}
	return config.jwtMac.ComputeMACAndEncode(rawJWT)
}

// VerifyAssumeJWTToken is the identity name and control admin of a link token for the site, refusing one for another
// site, expired, or already followed
func (config *SecurityConfig) VerifyAssumeJWTToken(tokenStr string, subdomain string) (string, string, error) {
	validator, err := jwt.NewValidator(&jwt.ValidatorOpts{ExpectedAudience: &AssumeAudience, ExpectIssuedInThePast: true, IgnoreIssuer: true})
	if err != nil {
		return "", "", err
	}
	verified, err := config.jwtMac.VerifyMACAndDecode(tokenStr, validator)
	if err != nil {
		return "", "", err
	}
	if issuer, err := verified.Issuer(); err != nil || issuer != subdomain {
		return "", "", fmt.Errorf("the link is for a different site")
	}
	name, err := verified.Subject()
	if err != nil {
		return "", "", err
	}
	by, err := verified.StringClaim("by")
	if err != nil {
		return "", "", err
	}
	id, err := verified.JWTID()
	if err != nil {
		return "", "", err
	}
	expiresAt, err := verified.ExpiresAt()
	if err != nil {
		return "", "", err
	}

	assumeUsed.Lock()
	defer assumeUsed.Unlock()
	for used, until := range assumeUsed.until {
		if until.Before(time.Now()) {
			delete(assumeUsed.until, used)
		}
	}
	if _, used := assumeUsed.until[id]; used {
		return "", "", fmt.Errorf("the link has been used")
	}
	assumeUsed.until[id] = expiresAt
	return name, by, nil
}
