//go:build ignore

package main

import (
	"bytes"
	"encoding/base64"
	"fmt"

	"log"

	"github.com/tink-crypto/tink-go/v2/aead"
	"github.com/tink-crypto/tink-go/v2/insecurecleartextkeyset"
	"github.com/tink-crypto/tink-go/v2/jwt"
	"github.com/tink-crypto/tink-go/v2/keyset"
	"github.com/tink-crypto/tink-go/v2/streamingaead"
)

func main() {
	// 1. Create the handle
	aesKS, err := createBase64AESGCMKeyset()
	if err != nil {
		log.Fatal(fmt.Errorf("failed to create AES-GCM keyset: %w", err))
	}

	streamKS, err := createBase64StreamingAEADKeyset()
	if err != nil {
		log.Fatal(fmt.Errorf("failed to create streaming AEAD keyset: %w", err))
	}

	jwtKS, err := createBase64JWTKeyset()
	if err != nil {
		log.Fatal(fmt.Errorf("failed to create JWT keyset: %w", err))
	}

	fmt.Printf("%s,%s,%s", aesKS, streamKS, jwtKS)
}

func createBase64StreamingAEADKeyset() (string, error) {
	// 1. Create the handle
	nkh, err := keyset.NewHandle(streamingaead.AES256GCMHKDF4KBKeyTemplate())
	if err != nil {
		log.Fatal(err)
	}
	return base64AKeySet(nkh)
}

func createBase64AESGCMKeyset() (string, error) {
	// 1. Create the handle
	nkh, err := keyset.NewHandle(aead.AES256GCMKeyTemplate())
	if err != nil {
		log.Fatal(err)
	}

	return base64AKeySet(nkh)
}

func createBase64JWTKeyset() (string, error) {
	// 1. Create the handle
	nkh, err := keyset.NewHandle(jwt.HS256Template())
	if err != nil {
		log.Fatal(err)
	}

	return base64AKeySet(nkh)
}

func base64AKeySet(handle *keyset.Handle) (string, error) {
	// 2. Export to Binary (for the Base64 string)
	binBuf := new(bytes.Buffer)
	binWr := keyset.NewBinaryWriter(binBuf)
	if err := insecurecleartextkeyset.Write(handle, binWr); err != nil {
		log.Fatal(err)
	}
	keySetString := base64.RawStdEncoding.EncodeToString(binBuf.Bytes())
	return keySetString, nil
}
