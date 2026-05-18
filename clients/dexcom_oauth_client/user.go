package dexcom_oauth_client

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

const (
	applicationId    = "d89443d2-327c-4a6f-89e5-496bbb0317db"
	agent            = "Dexcom Share/3.0.2.11 CFNetwork/711.2.23 Darwin/14.0.0"
	loginUrl         = "https://share1.dexcom.com/ShareWebServices/Services/General/LoginPublisherAccountByName"
	latestGlucoseUrl = "https://share1.dexcom.com/ShareWebServices/Services/Publisher/ReadPublisherLatestGlucoseValues"
)

func (client *DexcomClient) GetSessionID(username, password string) error {
	payload := map[string]string{
		"accountName":   username,
		"password":      password,
		"applicationId": applicationId,
	}
	payloadBytes, _ := json.Marshal(&payload)
	payloadReader := bytes.NewReader(payloadBytes)
	req, _ := http.NewRequest("POST", loginUrl, payloadReader)
	req.Header.Add("user-agent", agent)
	req.Header.Add("content-type", "application/json")
	req.Header.Add("accept", "application/json")
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return err
	}

	sessionId, _ := io.ReadAll(resp.Body)

	var id []byte
	// Strip quotations from the response
	for _, b := range sessionId {
		if b != 34 {
			id = append(id, b)
		}
	}

	client.DexcomToken = string(id)
	return nil
}
