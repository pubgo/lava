package golug_rabbitmq

import (
	"fmt"

	"github.com/pubgo/golug/golug_entry"
	"github.com/pubgo/golug/golug_plugin"
	"github.com/pubgo/xerror"
)

func init() {
	xerror.Exit(golug_plugin.Register(&golug_plugin.Base{
		Name: name,
		OnInit: func(ent golug_entry.Entry) {
			factories.Delete(name)
			deleteRabbitPool(name)

			//get or update redis client
			var rabbitConfig RabbitConfig
			xerror.Panic(ent.Decode(name, &rabbitConfig))
			xerror.Panic(initRabbitPool(name, &rabbitConfig))
		},
	}))
}

func StartRabbitWithFile(rbConfigs map[string]*RabbitConfig) error {
	if len(rbConfigs) == 0 {
		return fmt.Errorf("conf rabbit key prefix is empty")
	}
	for k, v := range rbConfigs {
		err := initRabbitPool(k, v)
		if err != nil {
			log.Errorf("BuildRedisData failed redisName %v, redisConfig: %v, err: %v", k, v, err)
			return err
		}
	}
	return nil
}
