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

func setup(t *testing.T) {
	err := godotenv.Load()
	if err != nil {
		t.Fatal("Error loading .env file")
	}
	manager.ConnectAPI()
}

func testCreateHelper(t *testing.T, data map[string]interface{}, expectedStatus int) string {
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

	if res.StatusCode == http.StatusCreated {
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
		return responseData.Message[index:]
	}
	return ""
}

func TestCreateEmptyRole(t *testing.T) {
	setup(t)
	data := map[string]interface{}{
		"email":    "__test100@example.com",
		"password": "Test123!",
	}
	uid := testCreateHelper(t, data, http.StatusCreated)
	manager.Auth0API.User.Delete(uid)
}

func TestCreateOnlyPengelola(t *testing.T) {
	setup(t)
	data := map[string]interface{}{
		"email":    "__test100@example.com",
		"password": "Test123!",
		"roles":    []string{"A1:Admin PPE", "A1:Admin Agency", "A1:Verifikator", "A1:Helpdesk", "A2:Admin PPE", "A2:Admin Agency", "A3:Verifikator", "A3:Helpdesk"},
	}
	uid := testCreateHelper(t, data, http.StatusCreated)
	manager.Auth0API.User.Delete(uid)
}

func TestCreateOnlyPengadaan(t *testing.T) {
	setup(t)
	data := map[string]interface{}{
		"email":    "__test100@example.com",
		"password": "Test123!",
		"roles":    []string{"A2:PPK", "A2:KUPBJ", "A2:Anggota Pokmil", "A3:PP", "A3:KUPBJ", "A3:Anggota Pokmil"},
	}
	uid := testCreateHelper(t, data, http.StatusCreated)
	manager.Auth0API.User.Delete(uid)
}

func TestCreateOnlyAuditor(t *testing.T) {
	setup(t)
	data := map[string]interface{}{
		"email":    "__test100@example.com",
		"password": "Test123!",
		"roles":    []string{"A1:Auditor", "A2:Auditor", "A3:Auditor"},
	}
	uid := testCreateHelper(t, data, http.StatusCreated)
	manager.Auth0API.User.Delete(uid)
}

func TestCreatePP_PPK(t *testing.T) {
	setup(t)
	data := map[string]interface{}{
		"email":    "__test100@example.com",
		"password": "Test123!",
		"roles":    []string{"A1:PPK", "A1:KUPBJ", "A1:Anggota Pokmil", "A1:PP"},
	}
	uid := testCreateHelper(t, data, http.StatusCreated)
	manager.Auth0API.User.Delete(uid)
}

func TestCreateCrossFunction(t *testing.T) {
	setup(t)
	data := map[string]interface{}{
		"email":    "__test100@example.com",
		"password": "Test123!",
		"roles":    []string{"A2:PPK", "A2:KUPBJ", "A2:Admin PPE"},
	}
	uid := testCreateHelper(t, data, http.StatusCreated)
	manager.Auth0API.User.Delete(uid)
}

func TestCreateCrossFunction2(t *testing.T) {
	setup(t)
	data := map[string]interface{}{
		"email":    "test100@example.com",
		"password": "Test123!",
		"roles":    []string{"A3:Helpdesk", "A3:Auditor"},
	}
	uid := testCreateHelper(t, data, http.StatusCreated)
	manager.Auth0API.User.Delete(uid)
}

// Extra Utility
func deleteUser(email string) error {
	userList, err := manager.Auth0API.User.List()
	if err != nil {
		return err
	}

	for _, user := range userList.Users {
		if *user.Email == email {
			manager.Auth0API.User.Delete(*user.ID)
			return nil
		}
	}
	return nil
}

func deleteAllTest() error {
	userList, err := manager.Auth0API.User.List()
	if err != nil {
		return err
	}

	uid_to_be_deleted := make([]string, 0)
	for _, user := range userList.Users {
		if strings.Contains(*user.Email, "test") {
			uid_to_be_deleted = append(uid_to_be_deleted, *user.ID)
		}
	}

	for _, uid := range uid_to_be_deleted {
		manager.Auth0API.User.Delete(uid)
	}
	return nil
}
