package gops

import (
	"github.com/google/gops/agent"

	"github.com/pubgo/funk/v2/log"
)

func init() {
	err := agent.Listen(agent.Options{})
	if err != nil {
		log.Err(err).Msg("failed to start gops agent")
	} else {
		log.Info().Msg("gops agent started ok")
	}
}
