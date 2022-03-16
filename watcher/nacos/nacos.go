package nacos

import (
	"context"
	"github.com/pubgo/lava/config/config_type"
	"github.com/pubgo/lava/watcher/watcher_type"

	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/model"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/pubgo/x/merge"
	"github.com/pubgo/x/strutil"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/clients/nacos"
	"github.com/pubgo/lava/event"
	watcher "github.com/pubgo/lava/watcher"
)

func init() {
	watcher.RegisterFactory(Name, func(cfg config_type.CfgMap) (watcher_type.Watcher, error) {
		var c Cfg
		xerror.Panic(merge.MapStruct(&c, cfg))
		return NewNacos(c)
	})
}

func NewNacos(cfg Cfg) (*nacosWatcher, error) {
	var c = nacos.Get(cfg.Driver)
	xerror.Assert(c == nil, "please init nacos client")
	manager := &nacosWatcher{client: c.GetCfg(), cfg: cfg}
	return manager, nil
}

var _ watcher_type.Watcher = (*nacosWatcher)(nil)

type nacosWatcher struct {
	cfgMap []vo.ConfigParam
	client config_client.IConfigClient
	cfg    Cfg
}

func (cm *nacosWatcher) Name() string { return Name }
func (cm *nacosWatcher) Close(ctx context.Context, opts ...watcher_type.Opt) {
	for i := range cm.cfgMap {
		_ = cm.client.CancelListenConfig(cm.cfgMap[i])
	}
}

func (cm *nacosWatcher) Get(ctx context.Context, group string, opts ...watcher_type.Opt) ([]*watcher_type.Response, error) {
	var cfgMap = xerror.PanicErr(cm.client.SearchConfig(vo.SearchConfigParam{
		Search:   "accurate", // 精确搜索
		Group:    group,
		PageSize: 1000,
	})).(*model.ConfigPage)

	var data = make([]*watcher_type.Response, len(cfgMap.PageItems))
	for i := range cfgMap.PageItems {
		var item = cfgMap.PageItems[i]
		data[i] = &watcher_type.Response{
			Event: event.EventType_UPDATE,
			Key:   item.DataId,
			Value: strutil.ToBytes(item.Content),
		}
	}
	return data, nil
}

func (cm *nacosWatcher) GetCallback(ctx context.Context, key string, fn func(resp *watcher_type.Response), opts ...watcher_type.Opt) error {
	var dt, err = cm.Get(ctx, key, opts...)
	if err != nil {
		return err
	}

	for i := range dt {
		fn(dt[i])
	}
	return nil
}

func (cm *nacosWatcher) Watch(ctx context.Context, dataId string, opts ...watcher_type.Opt) <-chan *watcher_type.Response {
	resp := make(chan *watcher_type.Response)
	configParams := vo.ConfigParam{
		DataId: dataId,
		Group:  cm.cfg.Group,
		OnChange: func(namespace, group, dataId, data string) {
			resp <- &watcher_type.Response{
				Key:   dataId,
				Value: strutil.ToBytes(data),
				Event: event.EventType_UPDATE,
			}
		},
	}
	xerror.Panic(cm.client.ListenConfig(configParams))
	cm.cfgMap = append(cm.cfgMap, configParams)
	return resp
}

func (cm *nacosWatcher) WatchCallback(ctx context.Context, key string, fn func(resp *watcher_type.Response), opts ...watcher_type.Opt) {
	for resp := range cm.Watch(ctx, key, opts...) {
		fn(resp)
	}
}