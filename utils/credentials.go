package utils

import (
	"fmt"

	"github.com/Ikit777/Andalalin-Backend/repository"
)

// GetCredentialsByRole func for getting credentials from a role name.
func GetCredentialsByRole(role string) ([]string, error) {
	// Define credentials variable.
	var credentials []string

	// Switch given role.
	switch role {
	case repository.SuperAdminRoleName:
		// Super Admin credentials.
		credentials = []string{
			repository.UserAddCredential,
			repository.UserDeleteCredential,
			repository.UserGetCredential,

			repository.AndalalinGetCredential,
			repository.AndalalinUpdateCredential,
			repository.AndalalinPersyaratanredential,
			repository.AndalalinStatusCredential,
			repository.AndalalinAddOfficerCredential,
			repository.AndalalinOfficerCredential,
			repository.AndalalinSurveyCredential,
			repository.AndalalinTicket1Credential,
			repository.AndalalinTicket2Credential,
			repository.AndalalinPersetujuanCredential,
			repository.AndalalinBAPCredential,
			repository.AndalalinSKCredential,
			repository.AndalalinKelolaTiket,
			repository.AndalalinKeputusanHasil,
			repository.AndalalinSurveiKepuasan,

			repository.ProductAddCredential,
			repository.ProductDeleteCredential,
			repository.ProductUpdateCredential,
		}
	case repository.DinasPerhubunganRoleName:
		// Admin credentials.
		credentials = []string{
			repository.AndalalinTicket1Credential,
			repository.AndalalinGetCredential,
			repository.AndalalinSurveyCredential,
			repository.AndalalinSurveiKepuasan,
		}
	case repository.AdminRoleName:
		// Admin credentials.
		credentials = []string{
			repository.AndalalinPersetujuanCredential,
			repository.AndalalinKelolaTiket,
			repository.AndalalinTicket2Credential,
			repository.AndalalinGetCredential,
			repository.AndalalinSurveyCredential,
			repository.AndalalinKeputusanHasil,
			repository.AndalalinSurveiKepuasan,

			repository.ProductAddCredential,
			repository.ProductDeleteCredential,
			repository.ProductUpdateCredential,
		}
	case repository.OperatorRoleName:
		// Operator credentials.
		credentials = []string{
			repository.AndalalinTicket1Credential,
			repository.AndalalinPersyaratanredential,
			repository.AndalalinStatusCredential,
			repository.AndalalinAddOfficerCredential,
			repository.AndalalinBAPCredential,
			repository.AndalalinSKCredential,
			repository.AndalalinSurveyCredential,
			repository.AndalalinGetCredential,
			repository.AndalalinKelolaTiket,
			repository.AndalalinSurveiKepuasan,
		}
	case repository.OfficerRoleName:
		// Officer credentials.
		credentials = []string{
			repository.AndalalinSurveyCredential,
			repository.AndalalinTicket2Credential,
			repository.AndalalinOfficerCredential,
			repository.AndalalinSurveiKepuasan,
			repository.AndalalinGetCredential,
		}
	case repository.UserRoleName:
		// User credentials.
		credentials = []string{
			repository.AndalalinPengajuanCredential,
			repository.AndalalinUpdateCredential,
		}
	default:
		// Return error message.
		return nil, fmt.Errorf("role '%v' does not exist", role)
	}

	return credentials, nil
}
