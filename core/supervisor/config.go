package supervisor

import (
	"context"
	"time"
)

func normalizeServiceConfig(config ServiceConfig) ServiceConfig {
	if config.RestartPolicy < RestartAlways || config.RestartPolicy > RestartNever {
		config.RestartPolicy = RestartAlways
	}

	if config.MaxRestarts < 0 {
		config.MaxRestarts = 0
	}

	if config.MaxRestartsInWindow < 0 {
		config.MaxRestartsInWindow = 0
	}

	if config.RestartDelay <= 0 {
		config.RestartDelay = time.Second
	}

	if config.MaxRestartDelay <= 0 || config.MaxRestartDelay < config.RestartDelay {
		config.MaxRestartDelay = config.RestartDelay
	}

	if config.RestartWindow <= 0 {
		config.RestartWindow = 5 * time.Minute
	}

	if config.BackoffMultiplier < 1 {
		config.BackoffMultiplier = 1
	}

	return config
}

func waitDelay(ctx context.Context, delay time.Duration) bool {
	if delay <= 0 {
		select {
		case <-ctx.Done():
			return false
		default:
			return true
		}
	}

	timer := time.NewTimer(delay)
	defer func() {
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
	}()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
