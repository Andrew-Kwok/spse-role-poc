package manager

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
)

type userInfo struct {
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
	w.Write([]byte(fmt.Sprintf(`{"message":"New user successfully creadted with ID: %s!"}`, *newUser.ID)))
}
