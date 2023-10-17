package utils

import (
	"fmt"

	"github.com/Ikit777/Andalalin-Backend/repository"
)

func GetRoleGives(role string) ([]string, error) {
	var rolegives []string

	// Switch given role.
	switch role {
	case repository.SuperAdminRoleName:
		// Super Admin credentials.
		rolegives = []string{
			repository.DinasPerhubunganRoleName,
			repository.AdminRoleName,
			repository.OperatorRoleName,
			repository.OfficerRoleName,
			repository.UserRoleName,
		}
	default:
		// Return error message.
		return nil, fmt.Errorf("role '%v' does not exist", role)
	}

	return rolegives, nil
}
