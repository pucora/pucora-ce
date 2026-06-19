package pucora

import (
	rss "github.com/pucora/pucora-rss/v2"
	xml "github.com/pucora/pucora-xml/v2"
	ginxml "github.com/pucora/pucora-xml/v2/gin"
	"github.com/pucora/lura/v2/router/gin"
)

// RegisterEncoders registers all the available encoders
func RegisterEncoders() {
	xml.Register()
	rss.Register()

	gin.RegisterRender(xml.Name, ginxml.Render)
}
