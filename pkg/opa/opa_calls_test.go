package opa

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

const fileContent = `externalSources:
  - name: "this-is-an-external-lib.rego"
    description: "This is an external lib"
    url: "%s/username/cc89a6225a301a35f0427b2bfc3db39b/raw/25a5d442a89026556729e9385e542ba936c083f5/rego1.rego"
  - name: "this-uses-the-lib.rego"
    description: "This uses the external lib"
    url: "%s/username/aa42f21e50e08a85e08ca0bd0ae91cec/raw/3ec2462919f949d6088995dc82f5b7e3fff099eb/rego2.txt"
    enabled: false
`

const check = `package fairwinds
hpaRequired[actionItem] {
  actionItem := {
    "title": "No horizontal pod autoscaler found"
  }
}
`

func TestGetExternalChecksFromFileWrongCreds(t *testing.T) {
	externalServer := httptest.NewServer(http.HandlerFunc(simpleExternalHandler))
	defer externalServer.Close()

	encodedAuth := base64.StdEncoding.EncodeToString([]byte("non-existing:non-existing"))
	header := "Authorization: Basic " + encodedAuth

	content := fmt.Sprintf(fileContent, externalServer.URL, externalServer.URL)
	c, err := getExternalChecksFromFile([]byte(content), []string{header})
	assert.EqualError(t, err, "error getting remote checks: invalid response code (401, expected 200)")
	assert.Len(t, c, 0)
}

func TestGetExternalChecksFromFile(t *testing.T) {
	externalServer := httptest.NewServer(http.HandlerFunc(simpleExternalHandler))
	defer externalServer.Close()

	encodedAuth := base64.StdEncoding.EncodeToString([]byte("username:password"))
	header := "Authorization: Basic " + encodedAuth

	content := fmt.Sprintf(fileContent, externalServer.URL, externalServer.URL)
	c, err := getExternalChecksFromFile([]byte(content), []string{header})
	assert.NoError(t, err)
	assert.Len(t, c, 2)
}

// simpleExternalHandler is a simple http handler that checks for basic auth and returns a check if the credentials are correct
func simpleExternalHandler(w http.ResponseWriter, r *http.Request) {
	user, pass, ok := r.BasicAuth()

	if !ok || !validateCredentials(user, pass) {
		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	fmt.Fprint(w, check, user)
}

func validateCredentials(username, password string) bool {
	return username == "username" && password == "password"
}
