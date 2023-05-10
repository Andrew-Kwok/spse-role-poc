package manager

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
)

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

	err = ValidateRoles(userinfo.Roles)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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

	if userinfo.Roles != nil && len(userinfo.Roles) > 0 {
		// generate a []management.Role array which contains the role to be assigned
		var RoleAssignment []*management.Role
		for _, role := range userinfo.Roles {
			role_obj, err := Auth0API.Role.Read(RoleID[role])
			if err != nil {
				Auth0API.User.Delete(*newUser.ID) // deletes the created user due to error
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			RoleAssignment = append(RoleAssignment, role_obj)
		}

		err = Auth0API.User.AssignRoles(*newUser.ID, RoleAssignment)
		if err != nil {
			Auth0API.User.Delete(*newUser.ID) // deletes the created user due to error
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf(`{"message":"New user successfully creadted with ID: %s"}`, *newUser.ID)))
}

func AddRolesHandler(w http.ResponseWriter, r *http.Request) {
	var userinfo userInfo
	err := json.NewDecoder(r.Body).Decode(&userinfo)

	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if userinfo.ID == "" {
		http.Error(w, "user id cannot be empty", http.StatusBadRequest)
		return
	}
	if userinfo.Roles == nil || len(userinfo.Roles) == 0 {
		http.Error(w, "To be added roles cannot be empty", http.StatusBadRequest)
		return
	}

	// get old roles for the current user, and check if the roles combined
	// with the future roles will trigger and error
	old_roles, err := Auth0API.User.Roles(userinfo.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	old_roles_name := []string{}
	for _, role := range old_roles.Roles {
		old_roles_name = append(old_roles_name, *role.Name)
	}

	err = ValidateRoles(append(old_roles_name, userinfo.Roles...))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// generate a []management.Role array which contains the role to be assigned
	var RoleAssignment []*management.Role
	for _, role := range userinfo.Roles {
		role_obj, err := Auth0API.Role.Read(RoleID[role])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		RoleAssignment = append(RoleAssignment, role_obj)
	}

	err = Auth0API.User.AssignRoles(userinfo.ID, RoleAssignment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"message":" Roles successfully added"}`)))
}

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
