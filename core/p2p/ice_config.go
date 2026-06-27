package p2p

import "github.com/pubgo/lava/v2/core/p2p/ice"

func toICEConfig(c Config, peerID string) (ice.Config, error) {
	turnURL := c.TURN.URL
	if c.TURN.Disabled {
		turnURL = ""
	}
	user, pass := c.TURN.Username, c.TURN.Password
	if turnURL != "" {
		var err error
		user, pass, err = c.TURN.Resolved(peerID)
		if err != nil {
			return ice.Config{}, err
		}
	}
	return ice.Config{
		STUNURLs:   c.STUNURLs,
		TURNURL:    turnURL,
		TURNUser:   user,
		TURNPass:   pass,
		ICETimeout: c.ICETimeout,
		AuthToken:  c.AuthToken,
	}, nil
}

func toCandidatePair(p ice.CandidatePairInfo) CandidatePairInfo {
	return CandidatePairInfo{
		LocalType:  p.LocalType,
		RemoteType: p.RemoteType,
		LocalAddr:  p.LocalAddr,
		RemoteAddr: p.RemoteAddr,
	}
}
