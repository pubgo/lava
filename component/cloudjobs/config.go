package cloudjobs

import (
	"time"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/lava/pkg/typex"
	"google.golang.org/protobuf/proto"
	yaml "gopkg.in/yaml.v3"
)

const DefaultPrefix = "acj"
const DefaultTimeout = 15 * time.Second
const DefaultMaxRetry = 3
const DefaultRetryBackoff = time.Second
const senderKey = "sender"

type Config struct {
	// Streams: nats stream config
	Streams map[string]*StreamConfig `yaml:"streams"`

	// Consumers: nats consumer config
	Consumers map[string]typex.YamlListType[*ConsumerConfig] `yaml:"consumers"`
}

type StreamConfig struct {
	// Storage jetstream.StorageType
	Storage string `yaml:"storage"`

	// Subjects stream subscribe subject, e.g. nvr.speaker.* without prefix
	Subjects typex.YamlListType[string] `yaml:"subjects"`
}

type ConsumerConfig struct {
	// Consumer name without prefix
	Consumer *string `yaml:"consumer"`

	// Stream name without prefix
	Stream string `yaml:"stream"`

	// Subjects config
	Subjects typex.YamlListType[*strOrJobConfig] `yaml:"subjects"`

	// Job config
	Job *JobConfig `yaml:"job"`
}

type JobConfig struct {
	// Name subject name
	Name *string `yaml:"name"`

	// Timeout job executor timeout, default: DefaultTimeout
	Timeout *time.Duration `yaml:"timeout"`

	// MaxRetry max retries, default: DefaultMaxRetry
	MaxRetry *int `yaml:"max_retries"`

	// RetryBackoff retry backoff, default: DefaultRetryBackoff
	RetryBackoff *time.Duration `yaml:"retry_backoff"`
}

type jobHandler struct {
	// job name
	name string

	// job handler
	handler func(ctx *Context, args proto.Message) error

	// job config
	cfg *JobConfig
}

type strOrJobConfig JobConfig

func (p *strOrJobConfig) UnmarshalYAML(value *yaml.Node) error {
	if value.IsZero() {
		return nil
	}

	switch value.Kind {
	case yaml.ScalarNode:
		var data string
		if err := value.Decode(&data); err != nil {
			return errors.WrapCaller(err)
		}

		*p = strOrJobConfig(JobConfig{Name: &data})
		return nil
	case yaml.MappingNode:
		var data JobConfig
		if err := value.Decode(&data); err != nil {
			return errors.WrapCaller(err)
		}

		*p = strOrJobConfig(data)
		return nil
	default:
		var val any
		assert.Exit(value.Decode(&val))
		return errors.Format("yaml kind type error,kind=%v data=%v", value.Kind, val)
	}
}
