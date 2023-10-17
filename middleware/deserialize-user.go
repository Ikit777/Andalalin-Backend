package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Ikit777/Andalalin-Backend/initializers"
	"github.com/Ikit777/Andalalin-Backend/models"
	"github.com/Ikit777/Andalalin-Backend/utils"
	"github.com/gin-gonic/gin"
)

type Tokens struct {
	Access string
}

func DeserializeUser() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var access_token string

		authorizationHeader := ctx.Request.Header.Get("Authorization")
		fields := strings.Fields(authorizationHeader)

		if len(fields) != 0 && fields[0] == "Bearer" {
			access_token = fields[1]
		}

		if access_token == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "fail", "message": "You are not logged in"})
			return
		}

		config, _ := initializers.LoadConfig()

		claim, err := utils.ValidateToken(access_token, config.AccessTokenPublicKey)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusFailedDependency, gin.H{"status": "fail", "message": err.Error()})
			return
		}

		var user models.User
		result := initializers.DB.First(&user, "id = ?", fmt.Sprint(claim.UserID))
		if result.Error != nil {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"status": "fail", "message": "the user belonging to this token no logger exists"})
			return
		}

		ctx.Set("currentUser", user)
		ctx.Set("accessUser", access_token)
		ctx.Next()
	}
}
