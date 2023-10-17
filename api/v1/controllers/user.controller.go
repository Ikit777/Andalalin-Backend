package controllers

import (
	"bytes"
	"image"
	_ "image/jpeg"
	"image/png"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Ikit777/Andalalin-Backend/initializers"
	"github.com/Ikit777/Andalalin-Backend/models"
	"github.com/Ikit777/Andalalin-Backend/repository"
	"github.com/Ikit777/Andalalin-Backend/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	_ "time/tzdata"
)

type UserController struct {
	DB *gorm.DB
}

func NewUserController(DB *gorm.DB) UserController {
	return UserController{DB}
}

func itemExists(array []string, item string) bool {
	for _, value := range array {
		if value == item {
			return true
		}
	}
	return false
}

func (ac *UserController) GetMe(ctx *gin.Context) {
	currentUser := ctx.MustGet("currentUser").(models.User)

	userResponse := &models.UserResponse{
		ID:        currentUser.ID,
		Name:      currentUser.Name,
		Email:     currentUser.Email,
		Role:      currentUser.Role,
		Photo:     currentUser.Photo,
		CreatedAt: currentUser.CreatedAt,
		UpdatedAt: currentUser.UpdatedAt,
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": userResponse})
}

func (ac *UserController) GetUserByEmail(ctx *gin.Context) {
	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.UserGetCredential]

	if !credential {
		// Return status 403 and permission denied error message.
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": true,
			"msg":   "Permission denied",
		})
		return
	}

	emailUser := ctx.Param("emailUser")

	var user models.User

	result := ac.DB.First(&user, "email = ?", emailUser)

	if result.Error != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"status": "fail", "message": "pengguna tidak ditemukan"})
		return
	}

	userResponse := &models.UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Role:      user.Role,
		Photo:     user.Photo,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": userResponse})
}

func (ac *UserController) GetUsers(ctx *gin.Context) {

	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.UserGetCredential]

	if !credential {
		// Return status 403 and permission denied error message.
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": true,
			"msg":   "Permission denied",
		})
		return
	}

	var users []models.User

	results := ac.DB.Find(&users)

	if results.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": results.Error})
		return
	} else {
		var respone []models.UserResponse

		for _, s := range users {
			respone = append(respone, models.UserResponse{
				ID:        s.ID,
				Name:      s.Name,
				Email:     s.Email,
				Role:      s.Role,
				Photo:     s.Photo,
				CreatedAt: s.CreatedAt,
				UpdatedAt: s.UpdatedAt,
			})
		}

		ctx.JSON(http.StatusOK, gin.H{"status": "success", "results": len(respone), "data": respone})
	}
}

func (ac *UserController) GetNotifikasi(ctx *gin.Context) {

	currentUser := ctx.MustGet("currentUser").(models.User)

	var notif []models.Notifikasi

	results := ac.DB.Find(&notif, "id_user = ?", currentUser.ID)

	if results.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": results.Error})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "results": len(notif), "data": notif})

}

func (ac *UserController) ClearNotifikasi(ctx *gin.Context) {

	currentUser := ctx.MustGet("currentUser").(models.User)

	results := ac.DB.Delete(&models.Notifikasi{}, "id_user = ?", currentUser.ID)

	if results.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": results.Error})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"status": "success"})
}

func (ac *UserController) GetUsersSortRole(ctx *gin.Context) {
	role := ctx.Param("role")

	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.UserGetCredential]

	if !credential {
		// Return status 403 and permission denied error message.
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": true,
			"msg":   "Permission denied",
		})
		return
	}

	var users []models.User

	results := ac.DB.Find(&users, "role = ?", role)

	if results.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": results.Error})
		return
	} else {
		var respone []models.UserResponse

		for _, s := range users {
			respone = append(respone, models.UserResponse{
				ID:        s.ID,
				Name:      s.Name,
				Email:     s.Email,
				Role:      s.Role,
				Photo:     s.Photo,
				CreatedAt: s.CreatedAt,
				UpdatedAt: s.UpdatedAt,
			})
		}

		ctx.JSON(http.StatusOK, gin.H{"status": "success", "results": len(respone), "data": respone})
	}
}

func (ac *UserController) GetPetugas(ctx *gin.Context) {
	config, _ := initializers.LoadConfig()

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.AndalalinAddOfficerCredential]

	if !credential {
		// Return status 403 and permission denied error message.
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": true,
			"msg":   "Permission denied",
		})
		return
	}

	var users []models.User

	results := ac.DB.Find(&users, "role = ?", "Petugas")

	if results.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": results.Error})
		return
	} else {
		var respone []models.UserResponse

		for _, s := range users {
			respone = append(respone, models.UserResponse{
				ID:    s.ID,
				Name:  s.Name,
				Email: s.Email,
				Photo: s.Photo,
			})
		}

		ctx.JSON(http.StatusOK, gin.H{"status": "success", "results": len(respone), "data": respone})
	}
}

// Add User
func (ac *UserController) Add(ctx *gin.Context) {
	var payload *models.UserAdd

	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	hashedPassword, err := utils.HashPassword(payload.Password)
	if err != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": err.Error()})
		return
	}

	config, _ := initializers.LoadConfig()

	currentUser := ctx.MustGet("currentUser").(models.User)
	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.UserAddCredential]

	if !credential {
		// Return status 403 and permission denied error message.
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": true,
			"msg":   "Permission denied",
		})
		return
	}

	roleGives, err := utils.GetRoleGives(currentUser.Role)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": true, "msg": err.Error()})
		return
	}

	roleExist := itemExists(roleGives, payload.Role)
	if !roleExist {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": true, "msg": "Permission denied"})
		return
	}

	loc, _ := time.LoadLocation("Asia/Singapore")
	now := time.Now().In(loc).Format("02-01-2006")
	filePath := "assets/default.png"
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Eror saat membaca file"})
		return
	}
	newUser := models.User{
		Name:      payload.Name,
		Email:     strings.ToLower(payload.Email),
		Password:  hashedPassword,
		Role:      payload.Role,
		Photo:     fileData,
		Verified:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	result := ac.DB.Create(&newUser)

	if result.Error != nil && strings.Contains(result.Error.Error(), "duplicate key value violates unique") {
		ctx.JSON(http.StatusConflict, gin.H{"status": "fail", "message": "Email sudah digunakan"})
		return
	} else if result.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": "Telah terjadi sesuatu"})
		return
	}

	userResponse := &models.UserResponse{
		ID:        newUser.ID,
		Name:      newUser.Name,
		Email:     newUser.Email,
		Role:      newUser.Role,
		CreatedAt: newUser.CreatedAt,
		UpdatedAt: newUser.UpdatedAt,
	}
	ctx.JSON(http.StatusCreated, gin.H{"status": "success", "data": userResponse})
}

// Delete User
func (ac *UserController) Delete(ctx *gin.Context) {
	var payload *models.Delete

	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	config, _ := initializers.LoadConfig()

	currentUser := ctx.MustGet("currentUser").(models.User)

	accessUser := ctx.MustGet("accessUser").(string)

	claim, error := utils.ValidateToken(accessUser, config.AccessTokenPublicKey)
	if error != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": error.Error()})
		return
	}

	credential := claim.Credentials[repository.UserDeleteCredential]

	if !credential {
		// Return status 403 and permission denied error message.
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": true,
			"msg":   "Permission denied",
		})
		return
	}

	roleGives, err := utils.GetRoleGives(currentUser.Role)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": true, "msg": err.Error()})
		return
	}

	roleExist := itemExists(roleGives, payload.Role)
	if !roleExist {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": true, "msg": "Permission denied"})
		return
	}

	result := ac.DB.Delete(&models.User{}, "id = ?", payload.ID)

	if result.Error != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": true, "msg": "Pengguna tidak ditemukan"})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"status": "success", "msg": "Pengguna berhasil dihapus"})
}

func (ac *UserController) ForgotPassword(ctx *gin.Context) {
	var payload *models.ForgotPasswordInput

	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	var user models.User
	result := ac.DB.First(&user, "email = ?", strings.ToLower(payload.Email))
	if result.Error != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Akun tidak terdaftar"})
		return
	}

	if !user.Verified {
		ctx.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "Pengguna belum melakukan verifikasi"})
		return
	}

	// Generate Verification Code
	resetToken := utils.Encode(6)
	user.ResetToken = resetToken
	loc, _ := time.LoadLocation("Asia/Singapore")
	user.ResetAt = time.Now().In(loc).Add(time.Minute * 5)
	ac.DB.Save(&user)

	emailData := utils.ResetPassword{
		Code:    resetToken,
		Subject: "Kode Reset Akun Andalalin Anda",
	}

	utils.SendEmailReset(user.Email, &emailData)

	fogotpasswordResponse := &models.ForgotPasswordRespone{
		ResetToken: user.ResetToken,
		ResetAt:    user.ResetAt,
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": fogotpasswordResponse})
}

func (ac *UserController) ResetPassword(ctx *gin.Context) {
	var payload *models.ResetPasswordInput
	resetToken := ctx.Params.ByName("resetToken")

	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	if payload.Password != payload.PasswordConfirm {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Password tidak sama"})
		return
	}

	hashedPassword, _ := utils.HashPassword(payload.Password)
	loc, _ := time.LoadLocation("Asia/Singapore")
	var updatedUser models.User
	result := ac.DB.First(&updatedUser, "reset_token = ? AND reset_at > ?", resetToken, time.Now().In(loc))
	if result.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "fail", "message": "Reset token kada luarsa"})
		return
	}

	now := time.Now().In(loc).Format("02-01-2006")

	updatedUser.Password = hashedPassword
	updatedUser.ResetToken = ""
	updatedUser.UpdatedAt = now
	ac.DB.Save(&updatedUser)

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "Reset akun berhasil"})
}

func (ac *UserController) UpdatePhoto(ctx *gin.Context) {
	file, err := ctx.FormFile("profile")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	uploadedFile, err := file.Open()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer uploadedFile.Close()

	imageFile, _, err := image.Decode(uploadedFile)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error decode image ": err.Error()})
		return
	}

	newWidth := 500
	newHeight := 500
	resizedImage := utils.ResizeImage(imageFile, newWidth, newHeight)

	var buf bytes.Buffer
	if err := png.Encode(&buf, resizedImage); err != nil {
		log.Fatal("Error encode image :", err)
		return
	}

	currentUser := ctx.MustGet("currentUser").(models.User)

	var user models.User

	result := ac.DB.Model(&user).Where("id = ?", currentUser.ID).Update("photo", buf.Bytes())
	if result.Error != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Akun tidak ditemukan"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "photo": buf.Bytes()})
}
