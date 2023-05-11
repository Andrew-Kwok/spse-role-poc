package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"spse-role-poc/api/manager"

	"github.com/joho/godotenv"
)

func testHelper(t *testing.T, data map[string]interface{}, expectedStatus int) {
	err := godotenv.Load()
	if err != nil {
		t.Fatal("Error loading .env file")
	}

	manager.ConnectAPI()
	// manager.RoleSetup()

	server := httptest.NewServer(http.HandlerFunc(manager.CreateUserHandler))
	defer server.Close()

	jsonData, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to marshal JSON data: %v", err)
	}

	req, err := http.NewRequest("POST", server.URL+"/create", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != expectedStatus {
		t.Fatalf("unexpected status code: got %d, want %d", res.StatusCode, http.StatusCreated)
	}

	// No new user is created, thus no deletion required.
	if expectedStatus != http.StatusCreated {
		return
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	var responseData struct {
		Message string `json:"message"`
	}
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON data: %v", err)
	}

	// each user_id begins with auth0|<uid>
	// look for the index and extract the uid
	index := strings.Index(responseData.Message, "auth0|")

	// delete the user to complete the testing
	t.Log(responseData.Message[index:])
	manager.Auth0API.User.Delete(responseData.Message[index:])
}

func TestCreateEmptyRole(t *testing.T) {
	data := map[string]interface{}{
		"email":    "test100@example.com",
		"password": "Test123!",
	}
	testHelper(t, data, http.StatusCreated)
}

func TestCreateOnlyPengelola(t *testing.T) {
	data := map[string]interface{}{
		"email":    "test100@example.com",
		"password": "Test123!",
		"roles":    []string{"A1:Admin PPE", "A1:Admin Agency", "A1:Verifikator", "A1:Helpdesk", "A2:Admin PPE", "A2:Admin Agency", "A3:Verifikator", "A3:Helpdesk"},
	}
	testHelper(t, data, http.StatusCreated)
}

func TestCreateOnlyPengadaan(t *testing.T) {
	data := map[string]interface{}{
		"email":    "test100@example.com",
		"password": "Test123!",
		"roles":    []string{"A2:PPK", "A2:KUPBJ", "A2:Anggota Pokmil", "A3:PP", "A3:KUPBJ", "A3:Anggota Pokmil"},
	}
	testHelper(t, data, http.StatusCreated)
}

func TestCreateOnlyAuditor(t *testing.T) {
	data := map[string]interface{}{
		"email":    "test100@example.com",
		"password": "Test123!",
		"roles":    []string{"A1:Auditor", "A2:Auditor", "A3:Auditor"},
	}
	testHelper(t, data, http.StatusCreated)
}

func TestCreatePP_PPK(t *testing.T) {
	data := map[string]interface{}{
		"email":    "test100@example.com",
		"password": "Test123!",
		"roles":    []string{"A1:PPK", "A1:KUPBJ", "A1:Anggota Pokmil", "A1:PP"},
	}
	testHelper(t, data, http.StatusCreated)
}

func TestCreateCrossFunction(t *testing.T) {
	data := map[string]interface{}{
		"email":    "test100@example.com",
		"password": "Test123!",
		"roles":    []string{"A2:PPK", "A2:KUPBJ", "A2:Admin PPE"},
	}
	testHelper(t, data, http.StatusCreated)
}

func TestCreateCrossFunction2(t *testing.T) {
	data := map[string]interface{}{
		"email":    "test100@example.com",
		"password": "Test123!",
		"roles":    []string{"A3:Helpdesk", "A3:Auditor"},
	}
	testHelper(t, data, http.StatusCreated)
}
