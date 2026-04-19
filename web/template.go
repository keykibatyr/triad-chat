package web

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Template struct {
	htmlTmpl *template.Template
}

func ParseFilesSys(pattern []string) (Template, error) {
	tpl, err := template.ParseFiles(pattern...)
	if err != nil {
		return Template{}, fmt.Errorf("parsing files %v", err)
	}

	return Template{
		htmlTmpl: tpl,
	}, nil
}

func Must(t Template, err error) Template {
	if err != nil {
		panic(err)
	}

	return t
}

func (t Template) ExecuteTemplate(c *gin.Context, data interface{}) {
	tpl, err := t.htmlTmpl.Clone()
	if err != nil {
		log.Printf("cloning: %v", err)
		c.String(http.StatusInternalServerError, "Could not create")
		return
	}

	var buf bytes.Buffer
	err = tpl.ExecuteTemplate(&buf, "base", data)
	if err != nil {
		log.Printf("executing the template: %v", err)
		c.String(http.StatusInternalServerError, "Could not execute the template")
		return
	}

	c.Data(http.StatusOK, "text/html; charset=utf-8", buf.Bytes())
}

