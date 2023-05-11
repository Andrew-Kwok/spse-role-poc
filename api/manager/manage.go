package manager

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
)

// User Information
// A role must follows the following format: "{satuan_kerja}:{role_function}", e.g. "A1:PP", "A2:Admin PPE"
type userInfo struct {
	ID       string   `json:"id"`
	Email    string   `json:"email"`
	Password string   `json:"password"`
	Roles    []string `json:"roles"`
}

func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	var userinfo userInfo
	err := json.NewDecoder(r.Body).Decode(&userinfo)

	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if userinfo.Email == "" {
		http.Error(w, "Email cannot be empty", http.StatusBadRequest)
		return
	}
	if userinfo.Password == "" {
		http.Error(w, "Password cannot be empty", http.StatusBadRequest)
		return
	}

	// setup user information
	newUser := &management.User{
		Connection: auth0.String("Username-Password-Authentication"),
		Email:      auth0.String(userinfo.Email),
		Password:   auth0.String(userinfo.Password),
	}

	// Create a new user
	err = Auth0API.User.Create(newUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	var err_list []error = nil
	if userinfo.Roles != nil && len(userinfo.Roles) > 0 {
		userinfo.ID = *newUser.ID
		err_list = RewriteRoles(userinfo)
	}

	// parse messages into json
	type message struct {
		Message string   `"json:message"`
		Errors  []string `"json:errors"`
	}

	var err_list_str []string
	for _, err := range err_list {
		err_list_str = append(err_list_str, err.Error())
	}

	json.NewEncoder(w).Encode(message{
		Message: fmt.Sprintf("New user successfully creaded with ID: %s", *newUser.ID),
		Errors:  err_list_str,
	})
}

func RewriteRoles(userinfo userInfo) []error {
	// Note: Role Existence is checked in RoleSetup(prefix)
	role_by_prefix := make(map[string][]string)
	for _, role := range userinfo.Roles {
		idx := strings.Index(role, ":")
		if idx == -1 {
			return []error{fmt.Errorf("Role %s is not in correct format", role)}
		}

		role_prefix, role_suffix := role[:idx], role[idx+1:]
		role_by_prefix[role_prefix] = append(role_by_prefix[role_prefix], role_suffix)
	}

	var err_accumulator []error = make([]error, 0)
	for prefix, roles_suffix := range role_by_prefix {
		err := RoleSetup(prefix)
		if err != nil {
			err_accumulator = append(err_accumulator, err)
			continue
		}

		err = ValidateRoles(prefix, roles_suffix)
		if err != nil {
			err_accumulator = append(err_accumulator, err)
			continue
		}

		// generate a []management.Role array which contains the role to be assigned
		retrieve_role_info_ok := true
		var RoleAssignment []*management.Role
		for _, role_suffix := range roles_suffix {
			role_obj, err := Auth0API.Role.Read(RoleID[prefix+":"+role_suffix])
			if err != nil {
				retrieve_role_info_ok = false
				err_accumulator = append(err_accumulator, err)
				break
			}
			RoleAssignment = append(RoleAssignment, role_obj)
		}

		if !retrieve_role_info_ok {
			continue
		}

		err = Auth0API.User.AssignRoles(userinfo.ID, RoleAssignment)
		if err != nil {
			err_accumulator = append(err_accumulator, err)
		}
	}

	if len(err_accumulator) == 0 {
		return nil
	}
	return err_accumulator
}

// func AddRolesHandler(w http.ResponseWriter, r *http.Request) {
// 	var userinfo userInfo
// 	err := json.NewDecoder(r.Body).Decode(&userinfo)

// 	if err != nil {
// 		http.Error(w, "Invalid request body", http.StatusBadRequest)
// 		return
// 	}
// 	if userinfo.ID == "" {
// 		http.Error(w, "user id cannot be empty", http.StatusBadRequest)
// 		return
// 	}
// 	if userinfo.Roles == nil || len(userinfo.Roles) == 0 {
// 		http.Error(w, "To be added roles cannot be empty", http.StatusBadRequest)
// 		return
// 	}

// 	// get old roles for the current user, and check if the roles combined
// 	// with the future roles will trigger and error
// 	old_roles, err := Auth0API.User.Roles(userinfo.ID)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	old_roles_name := []string{}
// 	for _, role := range old_roles.Roles {
// 		old_roles_name = append(old_roles_name, *role.Name)
// 	}

// 	err = ValidateRoles(append(old_roles_name, userinfo.Roles...))
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		return
// 	}

// 	// generate a []management.Role array which contains the role to be assigned
// 	var RoleAssignment []*management.Role
// 	for _, role := range userinfo.Roles {
// 		role_obj, err := Auth0API.Role.Read(RoleID[role])
// 		if err != nil {
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 			return
// 		}
// 		RoleAssignment = append(RoleAssignment, role_obj)
// 	}

// 	err = Auth0API.User.AssignRoles(userinfo.ID, RoleAssignment)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusOK)
// 	w.Write([]byte(fmt.Sprintf(`{"message":" Roles successfully added"}`)))
// }

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	var userinfo userInfo
	err := json.NewDecoder(r.Body).Decode(&userinfo)

	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = Auth0API.User.Delete(userinfo.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf(`{"message":"Successfully deleted user"`)))
}
