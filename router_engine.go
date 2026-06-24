package pucora

import (
	"net"
	"strings"

	"github.com/gin-gonic/gin"

	botdetector "github.com/pucora/pucora-botdetector/v2/gin"
	httpsecure "github.com/pucora/pucora-httpsecure/v2/gin"
	lua "github.com/pucora/pucora-lua/v2/router/gin"
	"github.com/pucora/lura/v2/config"
	luragin "github.com/pucora/lura/v2/router/gin"
)

type engineFactory struct{}

func (engineFactory) NewEngine(cfg config.ServiceConfig, opt luragin.EngineOptions) *gin.Engine {
	return NewEngine(cfg, opt)
}

func NewEngine(cfg config.ServiceConfig, opt luragin.EngineOptions) *gin.Engine {
	engine := luragin.NewEngine(cfg, opt)

	logPrefix := "[SERVICE: Gin]"
	if err := httpsecure.Register(cfg.ExtraConfig, engine); err != nil && err != httpsecure.ErrNoConfig {
		opt.Logger.Warning(logPrefix+"[HTTPsecure]", err)
	} else if err == nil {
		opt.Logger.Debug(logPrefix + "[HTTPsecure] Successfully loaded module")
	}

	lua.Register(opt.Logger, cfg.ExtraConfig, engine)

	botdetector.Register(cfg, opt.Logger, engine)

	RegisterVirtualHost(cfg, engine)

	return engine
}

func parseCIDRs(networks []string) []*net.IPNet {
	var result []*net.IPNet
	for _, network := range networks {
		if _, ipNet, err := net.ParseCIDR(network); err == nil {
			result = append(result, ipNet)
		} else if ip := net.ParseIP(network); ip != nil {
			result = append(result, &net.IPNet{IP: ip, Mask: net.CIDRMask(32, 32)})
		}
	}
	return result
}

func isTrustedProxy(ipStr string, trustedCIDRs []*net.IPNet) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}
	for _, cidr := range trustedCIDRs {
		if cidr.Contains(ip) {
			return true
		}
	}
	return false
}

func getRealClientIP(c *gin.Context, trustedCIDRs []*net.IPNet, headers []string) string {
	for _, headerName := range headers {
		if ip := c.GetHeader(headerName); ip != "" {
			parts := strings.Split(ip, ",")
			for i := len(parts) - 1; i >= 0; i-- {
				ip := strings.TrimSpace(parts[i])
				if ip != "" && !isTrustedProxy(ip, trustedCIDRs) {
					return ip
				}
			}
		}
	}
	return c.ClientIP()
}