package pucora

import (
	fastjson "github.com/pucora/pucora-fastjson"
	ginfastjson "github.com/pucora/pucora-fastjson/gin"
	rss "github.com/pucora/pucora-rss/v2"
	xml "github.com/pucora/pucora-xml/v2"
	ginxml "github.com/pucora/pucora-xml/v2/gin"
	"github.com/pucora/lura/v2/router/gin"
)

// RegisterEncoders registers all the available encoders
func RegisterEncoders() {
	xml.Register()
	rss.Register()
	fastjson.Register()

	gin.RegisterRender(xml.Name, ginxml.Render)
	gin.RegisterRender(fastjson.Name, ginfastjson.Render)
}
