package main

import (
	"log"
	"net/http"

	"github.com/Ikit777/Andalalin-Backend/controllers"
	"github.com/Ikit777/Andalalin-Backend/initializers"
	"github.com/Ikit777/Andalalin-Backend/routes"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var (
	server              *gin.Engine
	AuthController      controllers.AuthController
	AuthRouteController routes.AuthRouteController

	UserController      controllers.UserController
	UserRouteController routes.UserRouteController

	AndalalinController      controllers.AndalalinController
	AndalalinRouteController routes.AndalalinRouteController

	DataMasterController      controllers.DataMasterControler
	DataMasterRouteController routes.DataMasterRouteController
)

func init() {
	config, err := initializers.LoadConfig()
	if err != nil {
		log.Fatal("Could not load environment variables", err)
	}

	initializers.ConnectDB(&config)

	AuthController = controllers.NewAuthController(initializers.DB)
	AuthRouteController = routes.NewAuthRouteController(AuthController)

	UserController = controllers.NewUserController(initializers.DB)
	UserRouteController = routes.NewRouteUserController(UserController)

	AndalalinController = controllers.NewAndalalinController(initializers.DB)
	AndalalinRouteController = routes.NewRouteAndalalinController(AndalalinController)

	DataMasterController = controllers.NewDataMasterControler(initializers.DB)
	DataMasterRouteController = routes.NewDataMasterRouteController(DataMasterController)

	server = gin.Default()
}

func main() {
	config, err := initializers.LoadConfig()
	if err != nil {
		log.Fatal("Could not load environment variables", err)
	}

	corsConfig := cors.DefaultConfig()

	corsConfig.AllowOrigins = []string{"*"}
	corsConfig.AllowCredentials = true

	corsConfig.AllowHeaders = []string{"*"}
	corsConfig.AllowMethods = []string{"*"}

	server.Use(cors.New(corsConfig))

	router := server.Group("/api/v1")
	router.GET("/healthchecker", func(ctx *gin.Context) {
		message := "Welcome to andalalin"
		ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": message})
	})

	AuthRouteController.AuthRoute(router)
	UserRouteController.UserRoute(router)
	AndalalinRouteController.AndalalainRoute(router)
	DataMasterRouteController.DataMasterRoute(router)
	log.Fatal(server.Run(":" + config.ServerPort))
}
