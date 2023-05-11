package manager

import (
	"errors"
	"fmt"
)

var Hierarchy = map[string][]string{
	"Pengelola LPSE":        {"Admin PPE", "Admin Agency", "Verifikator", "Helpdesk"},
	"Pelaku Pengadaan LPSE": {"PPK", "KUPBJ", "Anggota Pokmil", "PP"},
	"Auditor":               {"Auditor"},
}

var Division map[string]string // Division maps each `role name` to its division (parent)
var RoleID map[string]string   // RoleID maps each `role name` to its `role id`

// Generate the value for `Division` and `RoleID`
func RoleSetup(prefix string) error {
	Division = make(map[string]string)
	for div, roles := range Hierarchy {
		for _, role := range roles {
			Division[role] = div
		}
	}

	// TODO: Fix Looping, only requires looping with prefix
	// Auth0API.Role.List() Returns a list of roles sorted by role_name.
	rolelist, err := Auth0API.Role.List()
	if err != nil {
		return err
	}

	RoleID = make(map[string]string)
	for _, role := range rolelist.Roles {
		RoleID[*role.Name] = *role.ID
	}

	return nil
}

func ValidateRoles(prefix string, roles []string) error {
	// no roles => no issue
	if roles == nil || len(roles) == 0 {
		return nil
	}

	// Division[role] must be the same for all role in roles
	var div string = ""
	for _, role := range roles {
		role_div, ok := Division[role]
		if !ok {
			return errors.New(fmt.Sprintf("Role not found: %s", prefix+":"+role))
		}
		if div == "" {
			div = role_div
		} else if div != role_div {
			return errors.New(fmt.Sprintf("User's roles in %s may not cross-function different division: %s, %s", prefix, div, role_div))
		}
	}

	// special case: a user cannot have PP and PPK at the same time
	if div == "Pelaku Pengadaan LPSE" {
		role_PP, role_PPK := false, false

		for _, role := range roles {
			if role == "PP" {
				role_PP = true
			}
			if role == "PPK" {
				role_PPK = true
			}
		}

		if role_PP && role_PPK {
			return errors.New(fmt.Sprintf("User's roles in %s may not contain PP and PPK at the same time", prefix))
		}
	}

	return nil
}
