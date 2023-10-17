package controllers

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Ikit777/Andalalin-Backend/initializers"
	"github.com/Ikit777/Andalalin-Backend/models"
	"github.com/Ikit777/Andalalin-Backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	_ "time/tzdata"
)

type AuthController struct {
	DB *gorm.DB
}

func NewAuthController(DB *gorm.DB) AuthController {
	return AuthController{DB}
}

// SignUp User
func (ac *AuthController) SignUp(ctx *gin.Context) {
	var payload *models.UserSignUp

	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	if payload.Password != payload.PasswordConfirm {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "fail", "message": "Password tidak sama"})
		return
	}

	hashedPassword, err := utils.HashPassword(payload.Password)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	loc, _ := time.LoadLocation("Asia/Singapore")
	now := time.Now().In(loc).Format("02-01-2006")
	verification_code := utils.Encode(6)

	filePath := "assets/default.png"
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Eror saat membaca file"})
		return
	}

	newUser := models.User{
		Name:             payload.Name,
		Email:            strings.ToLower(payload.Email),
		Password:         hashedPassword,
		Role:             "User",
		Photo:            fileData,
		Verified:         false,
		VerificationCode: verification_code,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	result := ac.DB.Create(&newUser)

	if result.Error != nil && strings.Contains(result.Error.Error(), "duplicate key value violates unique") {
		ctx.JSON(http.StatusConflict, gin.H{"status": "fail", "message": "Email sudah digunakan"})
		return
	} else if result.Error != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Telah terjadi sesuatu"})
		return
	}

	emailData := utils.Verification{
		Code:    verification_code,
		Name:    newUser.Name,
		Subject: "Kode Verifikasi Akun Andalalin Anda",
	}

	utils.SendEmailVerification(newUser.Email, &emailData)

	userResponse := &models.UserResponse{
		ID:        newUser.ID,
		Name:      newUser.Name,
		Email:     newUser.Email,
		Photo:     newUser.Photo,
		Role:      newUser.Role,
		CreatedAt: newUser.CreatedAt,
		UpdatedAt: newUser.UpdatedAt,
	}
	ctx.JSON(http.StatusCreated, gin.H{"status": "success", "data": userResponse})
}

func (ac *AuthController) ResendVerification(ctx *gin.Context) {
	var payload *models.User

	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	var user models.User
	result := ac.DB.First(&user, "email = ?", strings.ToLower(payload.Email))
	if result.Error != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Invalid email"})
		return
	}

	emailData := utils.Verification{
		Code:    user.VerificationCode,
		Name:    user.Name,
		Subject: "Kode Verifikasi Akun Andalalin Anda",
	}

	utils.SendEmailVerification(user.Email, &emailData)
	ctx.JSON(http.StatusCreated, gin.H{"status": "success", "messege": "Kirim ulang email verifikasi berhasil"})
}

func (ac *AuthController) VerifyEmail(ctx *gin.Context) {

	code := ctx.Params.ByName("verificationCode")

	var updatedUser models.User
	result := ac.DB.First(&updatedUser, "verification_code = ?", code)
	if result.Error != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Pengguna tidak ditemukan"})
		return
	}

	if updatedUser.Verified {
		ctx.JSON(http.StatusConflict, gin.H{"status": "fail", "message": "Akun andalalin sudah diverifikasi"})
		return
	}

	updatedUser.VerificationCode = ""
	updatedUser.Verified = true
	ac.DB.Save(&updatedUser)

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "Verifikasi akun andalalin berhasil"})
}

func (ac *AuthController) SignIn(ctx *gin.Context) {
	var payload *models.UserSignIn

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
		ctx.JSON(http.StatusForbidden, gin.H{"status": "fail", "message": "Akun belum melakukan verifikasi"})
		return
	}

	if err := utils.VerifyPassword(user.Password, payload.Password); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Akun tidak terdaftar"})
		return
	}

	credentials, err := utils.GetCredentialsByRole(user.Role)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	if user.Role == "User" || user.Role == "Operator" || user.Role == "Petugas" {
		if payload.PushToken != "" {
			result := ac.DB.Model(&user).Where("id = ?", user.ID).Update("push_token", payload.PushToken)
			if result.Error != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
				return
			}
		}
	}

	config, _ := initializers.LoadConfig()

	access_token, err := utils.CreateToken(config.AccessTokenExpiresIn, user.ID, config.AccessTokenPrivateKey, credentials)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	refresh_token, err := utils.CreateToken(config.RefreshTokenExpiresIn, user.ID, config.RefreshTokenPrivateKey, credentials)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	data := struct {
		Access    string    `json:"access_token,omitempty"`
		Refresh   string    `json:"refresh_token,omitempty"`
		Id        uuid.UUID `json:"id,omitempty"`
		Name      string    `json:"name,omitempty"`
		Email     string    `json:"email,omitempty"`
		Role      string    `json:"role,omitempty"`
		Photo     []byte    `json:"photo,omitempty"`
		PushToken string    `json:"push_token,omitempty"`
	}{
		Access:    access_token,
		Refresh:   refresh_token,
		Id:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Role:      user.Role,
		Photo:     user.Photo,
		PushToken: user.PushToken,
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": data})
}

func (ac *AuthController) RefreshAccessToken(ctx *gin.Context) {

	var refresh_token string

	authorizationHeader := ctx.Request.Header.Get("Authorization")
	fields := strings.Fields(authorizationHeader)

	if len(fields) != 0 && fields[0] == "Bearer" {
		refresh_token = fields[1]
	}

	config, _ := initializers.LoadConfig()

	claim, err := utils.ValidateToken(refresh_token, config.RefreshTokenPublicKey)
	if err != nil {
		getId := utils.GetIdByToken(refresh_token, config.AccessTokenPublicKey)
		var userData models.User
		initializers.DB.First(&userData, "id = ?", fmt.Sprint(getId.UserID))

		userData.PushToken = ""

		initializers.DB.Save(&userData)

		ctx.AbortWithStatusJSON(http.StatusFailedDependency, gin.H{"status": "fail", "message": err.Error()})
		return
	}

	expiresRefreshToken := claim.Expires

	loc, _ := time.LoadLocation("Asia/Singapore")

	now := time.Now().In(loc).Unix()

	if now < expiresRefreshToken {
		var user models.User
		result := initializers.DB.First(&user, "id = ?", fmt.Sprint(claim.UserID))
		if result.Error != nil {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"status": "fail", "message": "User tidak ditemukan"})
			return
		}
		credentials, err := utils.GetCredentialsByRole(user.Role)
		if err != nil {
			// Return status 400 and error message.
			ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": err.Error()})
			return
		}

		access_token, err := utils.CreateToken(config.AccessTokenExpiresIn, user.ID, config.AccessTokenPrivateKey, credentials)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"status": "fail", "message": err.Error()})
			return
		}

		ref_token, err := utils.CreateToken(config.RefreshTokenExpiresIn, user.ID, config.RefreshTokenPrivateKey, credentials)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"status": "fail", "message": err.Error()})
			return
		}

		data := struct {
			Access  string `json:"access_token,omitempty"`
			Refresh string `json:"refresh_token,omitempty"`
		}{
			Access:  access_token,
			Refresh: ref_token,
		}

		ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": data})
	} else {
		ctx.JSON(http.StatusUnauthorized, gin.H{"status": true, "msg": "Session telah berakhir"})
		return
	}
}

func (ac *AuthController) LogoutUser(ctx *gin.Context) {
	currentUser := ctx.MustGet("currentUser").(models.User)

	var user models.User
	result := ac.DB.First(&user, "id = ?", currentUser.ID)
	if result.Error != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": "fail", "message": "Akun tidak terdaftar"})
		return
	}

	user.PushToken = ""

	ac.DB.Save(&user)

	ctx.JSON(http.StatusOK, gin.H{"status": "success"})
}
