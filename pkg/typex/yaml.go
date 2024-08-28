package typex

import (
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/errors"
	yaml "gopkg.in/yaml.v3"
)

type YamlListType[T any] []T

func (p *YamlListType[T]) UnmarshalYAML(value *yaml.Node) error {
	if value.IsZero() {
		return nil
	}

	switch value.Kind {
	case yaml.ScalarNode, yaml.MappingNode:
		var data T
		if err := value.Decode(&data); err != nil {
			return errors.WrapCaller(err)
		}
		*p = []T{data}
		return nil
	case yaml.SequenceNode:
		var data []T
		if err := value.Decode(&data); err != nil {
			return errors.WrapCaller(err)
		}
		*p = data
		return nil
	default:
		var val any
		assert.Exit(value.Decode(&val))
		return errors.Format("yaml kind type error, kind=%v data=%v", value.Kind, val)
	}
}
