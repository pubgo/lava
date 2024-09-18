package cloudjobs

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/pubgo/funk/errors"
	"github.com/samber/lo"
)

func getStorageType(name string) jetstream.StorageType {
	switch name {
	case "memory":
		return jetstream.MemoryStorage
	case "file":
		return jetstream.FileStorage
	default:
		panic("unknown storage type")
	}
}

func mergeJobConfig(dst *JobConfig, src *JobConfig) *JobConfig {
	if src == nil {
		src = handleDefaultJobConfig(nil)
	}

	if dst == nil {
		dst = handleDefaultJobConfig(nil)
	}

	if dst.MaxRetry == nil {
		dst.MaxRetry = src.MaxRetry
	}

	if dst.Timeout == nil {
		dst.Timeout = src.Timeout
	}

	if dst.RetryBackoff == nil {
		dst.RetryBackoff = src.RetryBackoff
	}

	return dst
}

func handleDefaultJobConfig(cfg *JobConfig) *JobConfig {
	if cfg == nil {
		cfg = new(JobConfig)
	}

	if cfg.Timeout == nil {
		cfg.Timeout = lo.ToPtr(DefaultTimeout)
	}

	if cfg.MaxRetry == nil {
		cfg.MaxRetry = lo.ToPtr(DefaultMaxRetry)
	}

	if cfg.RetryBackoff == nil {
		cfg.RetryBackoff = lo.ToPtr(DefaultRetryBackoff)
	}

	return cfg
}

func handleSubjectName(name string, prefix string) string {
	prefix = fmt.Sprintf("%s.", prefix)
	if strings.HasPrefix(name, prefix) {
		return name
	}

	return fmt.Sprintf("%s%s", prefix, name)
}

func encodeDelayTime(duration time.Duration) string {
	return strconv.Itoa(int(time.Now().Add(duration).UnixMilli()))
}

func decodeDelayTime(delayTime string) (time.Duration, error) {
	tt, err := strconv.Atoi(delayTime)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to parse cloud job delay time, time=%s", delayTime)
	}
	return time.Until(time.UnixMilli(int64(tt))), nil
}
