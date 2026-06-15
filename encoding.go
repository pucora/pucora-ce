package velonetics

import (
	rss "github.com/velonetics/velonetics-rss/v2"
	xml "github.com/velonetics/velonetics-xml/v2"
	ginxml "github.com/velonetics/velonetics-xml/v2/gin"
	"github.com/velonetics/lura/v2/router/gin"
)

// RegisterEncoders registers all the available encoders
func RegisterEncoders() {
	xml.Register()
	rss.Register()

	gin.RegisterRender(xml.Name, ginxml.Render)
}
