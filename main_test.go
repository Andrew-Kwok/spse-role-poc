package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sort"
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

func testAddHelper(t *testing.T, data map[string]interface{}, expectedStatus int) {
	server := httptest.NewServer(http.HandlerFunc(manager.AddRolesHandler))
	defer server.Close()

	jsonData, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to marshal JSON data: %v", err)
	}

	req, err := http.NewRequest("POST", server.URL+"/addroles", bytes.NewBuffer(jsonData))
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
}

func checkRoles(t *testing.T, uid string, expectedRoles []string) error {
	rolelist, err := manager.Auth0API.User.Roles(uid)
	if err != nil {
		t.Fatal(err)
	}

	actualRoles := make([]string, 0)
	for _, role := range rolelist.Roles {
		actualRoles = append(actualRoles, *role.Name)
	}

	sort.Strings(actualRoles)
	sort.Strings(expectedRoles)

	if len(actualRoles) != len(expectedRoles) {
		t.Fatal("Expected ", expectedRoles, ". Got ", actualRoles)
	}

	for i := 0; i < len(expectedRoles); i++ {
		if actualRoles[i] != expectedRoles[i] {
			t.Fatal("Expected ", expectedRoles, ". Got ", actualRoles)
		}
	}
	return nil
}

func TestCreateEmptyRole(t *testing.T) {
	setup(t)
	data := map[string]interface{}{
		"email":    "__test100@example.com",
		"password": "Test123!",
	}
	uid := testCreateHelper(t, data, http.StatusCreated)
	defer manager.Auth0API.User.Delete(uid)
	checkRoles(t, uid, []string{})
}

func TestCreateOnlyPengelola(t *testing.T) {
	setup(t)
	data := map[string]interface{}{
		"email":    "__test100@example.com",
		"password": "Test123!",
		"roles":    []string{"A1:Admin PPE", "A1:Admin Agency", "A1:Verifikator", "A1:Helpdesk", "A2:Admin PPE", "A2:Admin Agency", "A3:Verifikator", "A3:Helpdesk"},
	}
	uid := testCreateHelper(t, data, http.StatusCreated)
	defer manager.Auth0API.User.Delete(uid)
	checkRoles(t, uid, []string{"A1:Admin PPE", "A1:Admin Agency", "A1:Verifikator", "A1:Helpdesk", "A2:Admin PPE", "A2:Admin Agency", "A3:Verifikator", "A3:Helpdesk"})
}

func TestCreateOnlyPengadaan(t *testing.T) {
	setup(t)
	data := map[string]interface{}{
		"email":    "__test100@example.com",
		"password": "Test123!",
		"roles":    []string{"A2:PPK", "A2:KUPBJ", "A2:Anggota Pokmil", "A3:PP", "A3:KUPBJ", "A3:Anggota Pokmil"},
	}
	uid := testCreateHelper(t, data, http.StatusCreated)
	defer manager.Auth0API.User.Delete(uid)
	checkRoles(t, uid, []string{"A2:PPK", "A2:KUPBJ", "A2:Anggota Pokmil", "A3:PP", "A3:KUPBJ", "A3:Anggota Pokmil"})
}

func TestCreateOnlyAuditor(t *testing.T) {
	setup(t)
	data := map[string]interface{}{
		"email":    "__test100@example.com",
		"password": "Test123!",
		"roles":    []string{"A1:Auditor", "A2:Auditor", "A3:Auditor"},
	}
	uid := testCreateHelper(t, data, http.StatusCreated)
	defer manager.Auth0API.User.Delete(uid)
	checkRoles(t, uid, []string{"A1:Auditor", "A2:Auditor", "A3:Auditor"})
}

func TestCreatePP_PPK(t *testing.T) {
	setup(t)
	data := map[string]interface{}{
		"email":    "__test100@example.com",
		"password": "Test123!",
		"roles":    []string{"A1:PPK", "A1:KUPBJ", "A1:Anggota Pokmil", "A1:PP"},
	}
	uid := testCreateHelper(t, data, http.StatusCreated)
	defer manager.Auth0API.User.Delete(uid)
	checkRoles(t, uid, []string{})
}

func TestCreateCrossFunction(t *testing.T) {
	setup(t)
	data := map[string]interface{}{
		"email":    "__test100@example.com",
		"password": "Test123!",
		"roles":    []string{"A2:PPK", "A2:KUPBJ", "A2:Admin PPE"},
	}
	uid := testCreateHelper(t, data, http.StatusCreated)
	defer manager.Auth0API.User.Delete(uid)
	checkRoles(t, uid, []string{})
}

func TestCreateCrossFunction2(t *testing.T) {
	setup(t)
	data := map[string]interface{}{
		"email":    "__test100@example.com",
		"password": "Test123!",
		"roles":    []string{"A1:Admin PPE", "A3:Helpdesk", "A3:Auditor"},
	}
	uid := testCreateHelper(t, data, http.StatusCreated)
	defer manager.Auth0API.User.Delete(uid)
	checkRoles(t, uid, []string{"A1:Admin PPE"})
}

func TestAddRoles(t *testing.T) {
	setup(t)
	data := map[string]interface{}{
		"email":    "__test100@example.com",
		"password": "Test123!",
	}
	uid := testCreateHelper(t, data, http.StatusCreated)
	defer manager.Auth0API.User.Delete(uid)

	data = map[string]interface{}{
		"id":    uid,
		"roles": []string{"A1:Admin PPE", "A1:Admin Agency", "A1:Verifikator", "A1:Helpdesk"},
	}
	testAddHelper(t, data, http.StatusOK)
	checkRoles(t, uid, []string{"A1:Admin PPE", "A1:Admin Agency", "A1:Verifikator", "A1:Helpdesk"})

	// No roles should be added
	data = map[string]interface{}{
		"id":    uid,
		"roles": []string{"A1:Admin PP"},
	}
	testAddHelper(t, data, http.StatusOK)
	checkRoles(t, uid, []string{"A1:Admin PPE", "A1:Admin Agency", "A1:Verifikator", "A1:Helpdesk"})

	// A2:Admin PP should be allowed to be added
	data = map[string]interface{}{
		"id":    uid,
		"roles": []string{"A2:PP"},
	}
	testAddHelper(t, data, http.StatusOK)
	checkRoles(t, uid, []string{"A1:Admin PPE", "A1:Admin Agency", "A1:Verifikator", "A1:Helpdesk", "A2:PP"})

	// A2:KUPBJ should be allowed to be added
	data = map[string]interface{}{
		"id":    uid,
		"roles": []string{"A2:KUPBJ"},
	}
	testAddHelper(t, data, http.StatusOK)
	checkRoles(t, uid, []string{"A1:Admin PPE", "A1:Admin Agency", "A1:Verifikator", "A1:Helpdesk", "A2:PP", "A2:KUPBJ"})

	// A2:PPK should not be allowed to be added since A2:PP exists
	data = map[string]interface{}{
		"id":    uid,
		"roles": []string{"A2:PPK"},
	}
	testAddHelper(t, data, http.StatusOK)
	checkRoles(t, uid, []string{"A1:Admin PPE", "A1:Admin Agency", "A1:Verifikator", "A1:Helpdesk", "A2:PP", "A2:KUPBJ"})

	// A2:Auditor should not be allowed to be added since it has functions in A2:Pelaku Pengadaan LPSE exists
	data = map[string]interface{}{
		"id":    uid,
		"roles": []string{"A2:Auditor"},
	}
	testAddHelper(t, data, http.StatusOK)
	checkRoles(t, uid, []string{"A1:Admin PPE", "A1:Admin Agency", "A1:Verifikator", "A1:Helpdesk", "A2:PP", "A2:KUPBJ"})

	// A3:Auditor should be allowed to be added since it has no functions in A3
	data = map[string]interface{}{
		"id":    uid,
		"roles": []string{"A3:Auditor"},
	}
	testAddHelper(t, data, http.StatusOK)
	checkRoles(t, uid, []string{"A1:Admin PPE", "A1:Admin Agency", "A1:Verifikator", "A1:Helpdesk", "A2:PP", "A2:KUPBJ", "A3:Auditor"})
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
