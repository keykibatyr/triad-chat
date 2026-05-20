package handler

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/keykibatyr/triad-chat/internal/repository"
	"github.com/keykibatyr/triad-chat/internal/utils"
)

type RegisterRequest struct {
	Email    string `form:"email" binding:"required"`
	Password string `form:"password" binding:"required,min=8"`
	Username string `form:"username" binding:"required"`
}

type LoginRequest struct {
	Email    string `form:"email" binding:"required"`
	Password string `form:"password" binding:"required,min=8"`
}

type AuthHandler struct {
	Templates struct {
		Register utils.Template
		Login    utils.Template
	}

	AuthService repository.AuthService
}

func (a *AuthHandler) Register(c *gin.Context) {
	data := gin.H{}
	utils.Render(c, a.Templates.Register, data)
	//do i need return here?
}

func (a *AuthHandler) RegisterProcess(c *gin.Context) {
	var regForm RegisterRequest

	err := c.ShouldBind(&regForm)
	if err != nil {
		utils.Render(c, a.Templates.Register, gin.H{
			"Error": "The Password or Email are Incorrect",
		})
		return
	}

	_, tokenPair, err := a.AuthService.Register(c, regForm.Email, regForm.Password, regForm.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register"})
		return
	}

	c.SetSameSite(http.SameSiteLaxMode)

	c.SetCookie(
		"refresh_token",
		tokenPair.RefreshToken,
		int(time.Until(tokenPair.RefreshExpiresAt).Seconds()),
		"/",
		"",
		false, //redo that, for now it is http, not https
		true,
	)

	c.SetCookie(
		"access_token",
		tokenPair.AccessToken,
		int(time.Until(tokenPair.AccessExpiresAt).Seconds()),
		"/",
		"",
		false,
		true,
	)
}

func (a *AuthHandler) Login(c *gin.Context) {
	// userID, role, err := middleware.CurrentUser(c)
	// log.Print(userID, role)
	// log.Print("[spweorjifobrgopw]")
	// if err == nil {
	// 	c.Redirect(http.StatusFound, "/")
	// 	return
	// }

	data := gin.H{}
	utils.Render(c, a.Templates.Login, data)
}

func (a *AuthHandler) LoginProcess(c *gin.Context) {
	var logForm LoginRequest

	err := c.ShouldBind(&logForm)
	if err != nil {
		panic(err)
	}

	_, tokenPair, err := a.AuthService.Login(c, logForm.Email, logForm.Password)
	if err != nil {
		panic(err)
	}

	c.SetSameSite(http.SameSiteLaxMode)

	c.SetCookie(
		"refresh_token",
		tokenPair.RefreshToken,
		int(time.Until(tokenPair.RefreshExpiresAt).Seconds()),
		"/",
		"",
		false, //redo that, for now it is http, not https
		true,
	)

	c.SetCookie(
		"access_token",
		tokenPair.AccessToken,
		int(time.Until(tokenPair.AccessExpiresAt).Seconds()),
		"/",
		"",
		false,
		true,
	)

	c.Redirect(http.StatusFound, "/chat")

}

func (a *AuthHandler) LogoutProcess(c *gin.Context) {
	tokenString, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register"})
		return
	}

	log.Print("bug1")

	err = a.AuthService.Logout(c, tokenString)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to logout"})
		return
	}

	c.SetCookie("access_token", "", -1, "/", "localhost", false, true)
	c.SetCookie("refresh_token", "", -1, "/", "localhost", false, true)

	c.Redirect(http.StatusFound, "/login")
}
