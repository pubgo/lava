package p2p

import "github.com/pubgo/lava/v2/core/p2p/ice"

func toICEConfig(c Config) ice.Config {
	return ice.Config{
		STUNURLs:   c.STUNURLs,
		TURNURL:    c.TURN.URL,
		TURNUser:   c.TURN.Username,
		TURNPass:   c.TURN.Password,
		ICETimeout: c.ICETimeout,
		AuthToken:  c.AuthToken,
	}
}

func toCandidatePair(p ice.CandidatePairInfo) CandidatePairInfo {
	return CandidatePairInfo{
		LocalType:  p.LocalType,
		RemoteType: p.RemoteType,
		LocalAddr:  p.LocalAddr,
		RemoteAddr: p.RemoteAddr,
	}
}
