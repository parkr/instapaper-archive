package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gomodule/oauth1/oauth"
	"github.com/ochronus/instapaper-go-client/instapaper"
)

const testEmailAddress = "test@example.com"
const testPassword = "testPassword123"
const testClientID = "client-id"
const testClientSecret = "client-secret"

func newTestInstapaperClient(emailAddress, password string, handler http.Handler) (*instapaper.Client, *httptest.Server, error) {
	server := httptest.NewServer(handler)
	apiClient := &instapaper.Client{
		OAuthClient: oauth.Client{
			SignatureMethod: oauth.HMACSHA1,
			Credentials: oauth.Credentials{
				Token:  testClientID,
				Secret: testClientSecret,
			},
			TokenRequestURI: server.URL + "/oauth/access_token",
		},
		Username: emailAddress,
		Password: password,
		BaseURL:  server.URL,
	}
	authErr := apiClient.Authenticate()
	if authErr != nil {
		return nil, server, fmt.Errorf("error authenticating: %v", authErr)
	}
	return apiClient, server, nil
}

func fileContentsMatch(t *testing.T, path, expected string) {
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("unable to open file: %v", err)
	}
	contents, err := ioutil.ReadAll(f)
	if err != nil {
		t.Fatalf("unable to read file: %v", err)
	}
	if !strings.Contains(string(contents), expected) {
		t.Fatalf("file %q does not contain %q:\n\n%s\n---", path, expected, string(contents))
	}
}
