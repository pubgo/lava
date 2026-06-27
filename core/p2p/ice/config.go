package ice

import "time"

// Role ICE 协商角色。
type Role int

const (
	RoleDialer Role = iota
	RoleListener
)

// Config ICE 客户端配置。
type Config struct {
	STUNURLs   []string
	TURNURL    string
	TURNUser   string
	TURNPass   string
	ICETimeout time.Duration
	AuthToken  string
}

// CandidatePairInfo ICE 选路摘要。
type CandidatePairInfo struct {
	LocalType  string
	RemoteType string
	LocalAddr  string
	RemoteAddr string
}
