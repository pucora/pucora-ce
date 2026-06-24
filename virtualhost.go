package pucora

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/pucora/lura/v2/config"
	"github.com/pucora/lura/v2/core"
	"github.com/pucora/lura/v2/transport/http/server"
)

const (
	virtualHostNamespace = "server/virtualhost"
)

type virtualHostConfig struct {
	Hosts         []string          `json:"hosts"`
	AliasedHosts  map[string]string `json:"aliased_hosts"`
}

func virtualHostMiddleware(cfg virtualHostConfig) gin.HandlerFunc {
	hostSet := make(map[string]bool)
	for _, h := range cfg.Hosts {
		hostSet[h] = true
	}

	return func(c *gin.Context) {
		host := c.GetHeader("Host")
		if host == "" {
			c.Next()
			return
		}

		if cfg.AliasedHosts != nil {
			if alias, ok := cfg.AliasedHosts[host]; ok {
				host = alias
			}
		}

		if !hostSet[host] {
			c.Next()
			return
		}

		originalPath := c.Request.URL.Path
		if strings.HasPrefix(originalPath, "/__virtual/") {
			c.Next()
			return
		}

		newPath := "/__virtual/" + host + originalPath
		c.Request.URL.Path = newPath
		c.Request.URL.RawPath = ""

		c.Header(core.PucoraHeaderName, core.PucoraHeaderValue)
		c.Header(server.CompleteResponseHeaderName, server.HeaderIncompleteResponseValue)

		c.Next()
	}
}

func getVirtualHostConfig(extraConfig map[string]interface{}) *virtualHostConfig {
	v, ok := extraConfig[virtualHostNamespace]
	if !ok || v == nil {
		return nil
	}

	cfg, ok := v.(map[string]interface{})
	if !ok {
		return nil
	}

	result := &virtualHostConfig{
		Hosts:        []string{},
		AliasedHosts: make(map[string]string),
	}

	if hosts, ok := cfg["hosts"].([]interface{}); ok {
		for _, h := range hosts {
			if s, ok := h.(string); ok {
				result.Hosts = append(result.Hosts, s)
			}
		}
	}

	if aliased, ok := cfg["aliased_hosts"].(map[string]interface{}); ok {
		for k, v := range aliased {
			if s, ok := v.(string); ok {
				result.AliasedHosts[k] = s
			}
		}
	}

	if len(result.Hosts) == 0 && len(result.AliasedHosts) == 0 {
		return nil
	}

	return result
}

func RegisterVirtualHost(cfg config.ServiceConfig, engine *gin.Engine) {
	vhConfig := getVirtualHostConfig(cfg.ExtraConfig)
	if vhConfig == nil {
		return
	}

	engine.Use(virtualHostMiddleware(*vhConfig))
}