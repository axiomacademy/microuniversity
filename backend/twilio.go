package main

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
)

var TWILIO_API_ERROR error = errors.New("Twilio API Error")

type TwilioApi struct {
	HttpClient    *http.Client
	AccountSid    string
	AuthToken     string
	VerifyBaseUrl string
}

func NewTwilioApi(AccountSid string, AuthToken string, VerifySid string) *TwilioApi {
	return &TwilioApi{
		HttpClient:    &http.Client{},
		AccountSid:    AccountSid,
		AuthToken:     AuthToken,
		VerifyBaseUrl: "https://verify.twilio.com/v2/Services/" + VerifySid,
	}
}

func (api *TwilioApi) StartEmailVerification(email string) error {
	reqData := url.Values{}
	reqData.Set("To", email)
	reqData.Set("Channel", "email")
	reqDataReader := strings.NewReader(reqData.Encode())

	// Creating the request
	req, _ := http.NewRequest("POST", api.VerifyBaseUrl+"/Verifications", reqDataReader)
	api.preflightRequest(req)

	resp, _ := api.HttpClient.Do(req)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// A successfull request
		return nil
	} else {
		return TWILIO_API_ERROR
	}
}

func (api *TwilioApi) VerifyCode(code string, email string) error {
	reqData := url.Values{}
	reqData.Set("To", email)
	reqData.Set("Code", code)
	reqDataReader := strings.NewReader(reqData.Encode())

	// Creating the request
	req, _ := http.NewRequest("POST", api.VerifyBaseUrl+"/VerificationCheck", reqDataReader)
	api.preflightRequest(req)

	resp, _ := api.HttpClient.Do(req)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// A successfull request
		return nil
	} else {
		// TODO: Determine the cause of the error, invalid code etc.
		return TWILIO_API_ERROR
	}
}

func (api *TwilioApi) preflightRequest(req *http.Request) {
	req.SetBasicAuth(api.AccountSid, api.AuthToken)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
}
