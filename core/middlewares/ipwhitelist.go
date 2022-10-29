package middlewares

import (
	"net"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

// IPWhitelist takes an acl map of IPs and checks incoming requests for matches.
// In acl map, the key is either an IP address or an IP network,
// while the bool value indicates it is allowed or blocked.
func IPWhitelist(acl map[string]bool) fiber.Handler {
	var allowedNets, blockedNets []*net.IPNet

	for s, allowed := range acl {
		_, ipNet, err := net.ParseCIDR(s)
		if err != nil {
			// might be an IP, keep it as is in acl
			continue
		}

		delete(acl, s)
		if allowed {
			allowedNets = append(allowedNets, ipNet)
		} else {
			blockedNets = append(blockedNets, ipNet)
		}
	}
	return func(ctx *fiber.Ctx) error {
		ip := ctx.IP()
		if acl == nil || len(acl) == 0 {
			return &fiber.Error{
				Code:    http.StatusForbidden,
				Message: "White list is not enabled, please contact the administrator",
			}
		}

		if allowed, ok := acl[ip]; ok {
			if !allowed {
				return &fiber.Error{
					Code:    http.StatusForbidden,
					Message: "The IP is forbidden, please contact the administrator",
				}
			}
			return nil
		}

		ipAddr := net.ParseIP(ip)
		for _, ipNet := range blockedNets {
			if ipNet.Contains(ipAddr) {
				acl[ip] = false
				return &fiber.Error{
					Code:    http.StatusForbidden,
					Message: "The IP is in the blacklist, please contact the administrator",
				}
			}
		}

		for _, ipNet := range allowedNets {
			if ipNet.Contains(ipAddr) {
				acl[ip] = true
				return ctx.Next()
			}
		}

		return &fiber.Error{
			Code:    http.StatusForbidden,
			Message: "The IP is denied, please contact the administrator",
		}
	}
}
