package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/keykibatyr/triad-chat/internal/auth"
	chatRepository "github.com/keykibatyr/triad-chat/internal/chat/repository"
	chatService "github.com/keykibatyr/triad-chat/internal/chat/service"
	"github.com/keykibatyr/triad-chat/internal/chat/worker"
	"github.com/keykibatyr/triad-chat/internal/config"
	"github.com/keykibatyr/triad-chat/internal/handler"
	"github.com/keykibatyr/triad-chat/internal/middleware"
	"github.com/keykibatyr/triad-chat/internal/repository"
	"github.com/keykibatyr/triad-chat/internal/service"
	"github.com/keykibatyr/triad-chat/internal/ws"
	"github.com/keykibatyr/triad-chat/web"
)

func main() {

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	cfg := config.Load()

	db, err := repository.Open(cfg.Database)
	if err != nil {
		log.Fatal("Error opening the db")
	}

	defer db.Close()

	err = repository.Migrate(db, "./migrations")
	if err != nil {
		log.Fatal("could not migrate tables")
	}

	fileUser := []string{
		"web/layout/base.tmpl",
		"web/partials/navbar.tmpl",
		"web/partials/footer.tmpl",
	}

	homePage := web.Must(web.ParseFilesSys(append(fileUser, "web/home/home.tmpl")))
	registerPage := web.Must(web.ParseFilesSys(append(fileUser, "web/auth/register.tmpl")))
	loginPage := web.Must(web.ParseFilesSys(append(fileUser, "web/auth/login.tmpl")))

	chat := web.Must(web.ParseFilesSys(append(fileUser, "web/public/chat.tmpl")))

	jwt := auth.NewJWTService(cfg.JWT.AccessTokenSecret, cfg.JWT.RefreshTokenSecret, cfg)

	userRepo := repository.NewUserRepo(db)
	authService := service.NewAuthService(userRepo, repository.NewRefreshRepo(db), jwt)

	authHandler := handler.AuthHandler{
		AuthService: authService,
	}

	authHandler.Templates.Register = registerPage
	authHandler.Templates.Login = loginPage

	authMiddleware := &middleware.AuthMiddleware{
		JWT:         jwt,
		AuthService: authService,
	}

	messageRepo := chatRepository.NewMessageRepo(db)
	convoRepo := chatRepository.NewConvoRepo(db)

	messageService := chatService.NewMessageService(convoRepo, messageRepo)

	persistWorker := worker.NewPersistWorker(messageService)
	hub := ws.NewHub(persistWorker)
	hub.Worker.SaveToDB(ctx)
	hubHandler := ws.NewHubHandler(hub, service.NewUserService(userRepo))

	hubHandler.Templates.Chat = chat

	router := gin.Default()

	router.GET("/register", authHandler.Register)
	router.POST("/register", authHandler.RegisterProcess)
	router.GET("/login", authHandler.Login)
	router.POST("/login", authHandler.LoginProcess)
	router.POST("/logout", authHandler.LogoutProcess)

	router.GET("/", func(c *gin.Context) {
		homePage.ExecuteTemplate(c, gin.H{})
	})

	protected := router.Group("/")
	{
		protected.Use(authMiddleware.AuthAccess())
		protected.GET("/ws", hubHandler.ServeWs)
		protected.GET("/chat", hubHandler.ChatPage)
		protected.POST("/ws/createRoom", hubHandler.CreateRoom)
	}

	router.Run()
}
