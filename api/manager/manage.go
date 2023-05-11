package manager

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
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

type message struct {
	Message string   `"json:message"`
	Errors  []string `"json:errors"`
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

func RewriteRolesHandler(w http.ResponseWriter, r *http.Request) {
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

	old_roles, err := Auth0API.User.Roles(userinfo.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	Auth0API.User.RemoveRoles(userinfo.ID, old_roles.Roles)
	err_list := RewriteRoles(userinfo)

	// parse messages into json
	var err_list_str []string
	for _, err := range err_list {
		err_list_str = append(err_list_str, err.Error())
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(message{
		Message: "Role Rewriting Successfully Halted. Please pay attention to the \"errors\" scope. You can ignore this message if it is empty.",
		Errors:  err_list_str,
	})

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
	sort.Strings(userinfo.Roles)

	// get old roles for the current user, and check if the roles combined
	// with the future roles will trigger an error
	// Only add old roles that has at least one common "satuan_kerja" as userinfo.Roles
	// Note: old_roles is sorted by role's Name
	old_roles, err := Auth0API.User.Roles(userinfo.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	left_bound := 0
	for i, role := range userinfo.Roles {
		idx := strings.Index(role, ":")
		if idx == -1 {
			continue
		}
		satKer := role[:idx]
		if i > 0 && strings.HasPrefix(userinfo.Roles[i-1], satKer+":") {
			// the role is already added in the previous iteration
			continue
		} else {
			left, right := left_bound, len(old_roles.Roles)-1
			for left < right {
				mid := (left + right) >> 1
				if *old_roles.Roles[mid].Name < satKer+":" {
					left = mid + 1
				} else {
					right = mid
				}
			}
			for ; left < len(old_roles.Roles) && strings.HasPrefix(*old_roles.Roles[left].Name, satKer+":"); left++ {
				userinfo.Roles = append(userinfo.Roles, *old_roles.Roles[left].Name)
			}
			// Since both old_roles and userinfo are sorted, future iterations on userinfo
			// must be in at the greater index.
			left_bound = left
		}
	}

	err_list := RewriteRoles(userinfo)

	// parse messages into json
	var err_list_str []string
	for _, err := range err_list {
		err_list_str = append(err_list_str, err.Error())
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(message{
		Message: "Role Addition Successfully Halted. Please pay attention to the \"errors\" scope. You can ignore this message if it is empty.",
		Errors:  err_list_str,
	})
}

func QueryAssignHandler(w http.ResponseWriter, r *http.Request) {
	type queryVar struct {
		AssignerUID string `json:"assigner_uid"`
		CreateRole  string `json:"create_role"`
	}
	var query queryVar
	err := json.NewDecoder(r.Body).Decode(&query)

	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if query.CreateRole == "" {
		http.Error(w, "Assignee UID cannot be empty", http.StatusBadRequest)
		return
	}

	idx := strings.Index(query.CreateRole, ":")
	if idx == -1 {
		http.Error(w, fmt.Sprintf("Role %s is not in correct format", query.CreateRole), http.StatusBadRequest)
	}

	satuanKerja, assigneeRole := query.CreateRole[:idx], query.CreateRole[idx+1:]
	assignerRolelist, err := Auth0API.User.Roles(query.AssignerUID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	left, right := 0, len(assignerRolelist.Roles)-1
	for left < right {
		mid := (left + right) >> 1
		if *assignerRolelist.Roles[mid].Name < satuanKerja+":" {
			left = mid + 1
		} else {
			right = mid
		}
	}

	// only PPE and Agency has the power to create other users
	// PPE can create all but PPE and Auditor
	// Agency can create all but PPE, Auditor, Agency
	assignerPPE, assignerAgency := false, false
	for ; left < len(assignerRolelist.Roles) && strings.HasPrefix(*assignerRolelist.Roles[left].Name, satuanKerja+":"); left++ {
		if *assignerRolelist.Roles[left].Name == satuanKerja+":"+"Admin PPE" {
			assignerPPE = true
		}
		if *assignerRolelist.Roles[left].Name == satuanKerja+":"+"Admin Agency" {
			assignerAgency = true
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if assignerPPE {
		if assigneeRole == "Admin PPE" || assigneeRole == "Auditor" {
			w.Write([]byte(`"message": "Action not allowed"`))
		} else {
			w.Write([]byte(`"message": "Action allowed"`))
		}
	} else if assignerAgency {
		if assigneeRole == "Admin PPE" || assigneeRole == "Auditor" || assigneeRole == "Admin Agency" {
			w.Write([]byte(`"message": "Action not allowed"`))
		} else {
			w.Write([]byte(`"message": "Action allowed"`))
		}
	} else {
		w.Write([]byte(`"message": "Action not allowed"`))
	}
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
