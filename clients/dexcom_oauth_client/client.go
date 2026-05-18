package dexcom_oauth_client

import (
	"fmt"
)

type DexcomClient struct {
	AuthCode    string
	DexcomToken string
	config      *Config
	oAuthToken  *Token
}

func NewClient(config *Config) *DexcomClient {
	dc := &DexcomClient{
		config: config,
	}

	if config.IsDev {
		fmt.Println("Dev server starting on :8000")
		url := config.getBaseUrl() + "/v1/oauth2/login?client_id=" + config.ClientId + "&redirect_uri=" + config.RedirectURI + "&response_type=code&scope=offline_access"
		fmt.Println(url)

	}
	return dc
}

func NewClientWithToken(config *Config, token *Token) *DexcomClient {
	return &DexcomClient{
		config:     config,
		oAuthToken: token,
	}
}
