package utils

import (
	"github.com/gin-gonic/gin"
	"github.com/keykibatyr/triad-chat/internal/middleware"
)

type Template interface {
	ExecuteTemplate(c *gin.Context, data interface{})
}

func Render(c *gin.Context, tpl Template, data gin.H) {
	if data == nil {
		data = gin.H{}
	}

	userID, _, err := middleware.CurrentUser(c) 
	if err != nil {
		data["CurrentUser"] = nil
		tpl.ExecuteTemplate(c, data)
		return
	}


	data["CurrentUser"] = userID

	tpl.ExecuteTemplate(c, data)
}
