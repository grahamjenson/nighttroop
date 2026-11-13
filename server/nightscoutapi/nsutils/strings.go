package nsutils

// ValidateIdentifierIsntBad checks if the identifier is valid and returns an error if it is not.
// Supports up to UUID 36 chars, alpha numeric and _ and - only
// This is because users are allowed to set their own identifiers, but we have limits
func ValidateIdentifierIsntBad(identifier string) *HTTPError {
	return validateIdentifier(identifier, MaxIdentifierLength)
}

// MaxIdentifierLength is the longest identifier a document can have
const MaxIdentifierLength = 36

// ValidateSettingKeyIsntBad checks a settings document's key, which is longer than an identifier: AAPS names its
// client control messages aaps_clientcontrol_ with their kind and the client's UUID, as
// aaps_clientcontrol_cmd_BolusPrepare_<uuid>. The key column keeps up to 128.
func ValidateSettingKeyIsntBad(key string) *HTTPError {
	return validateIdentifier(key, 128)
}

func validateIdentifier(identifier string, maxLength int) *HTTPError {
	if len(identifier) == 0 || len(identifier) > maxLength {
		return BadIdentifierError
	}

	// Below code is basically this, but just faster
	// if !regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(identifier) {
	// 	return BadIdentifierError
	// }

	for i := 0; i < len(identifier); i++ {
		c := identifier[i]
		if (c >= 'a' && c <= 'z') ||
			(c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') ||
			c == '_' ||
			c == '-' {
			continue
		}
		return BadIdentifierError
	}
	return nil
}
