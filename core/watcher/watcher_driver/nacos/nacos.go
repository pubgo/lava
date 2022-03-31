package nacos

import (
	"context"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/model"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/pubgo/lava/config"
	watcher3 "github.com/pubgo/lava/core/watcher"
	"github.com/pubgo/x/merge"
	"github.com/pubgo/x/strutil"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/clients/nacos"
	"github.com/pubgo/lava/event"
)

func init() {
	watcher3.RegisterFactory(Name, func(cfg config.CfgMap) (watcher3.Watcher, error) {
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

var _ watcher3.Watcher = (*nacosWatcher)(nil)

type nacosWatcher struct {
	cfgMap []vo.ConfigParam
	client config_client.IConfigClient
	cfg    Cfg
}

func (cm *nacosWatcher) Name() string { return Name }
func (cm *nacosWatcher) Close(ctx context.Context, opts ...watcher3.Opt) {
	for i := range cm.cfgMap {
		_ = cm.client.CancelListenConfig(cm.cfgMap[i])
	}
}

func (cm *nacosWatcher) Get(ctx context.Context, group string, opts ...watcher3.Opt) ([]*watcher3.Response, error) {
	var cfgMap = xerror.PanicErr(cm.client.SearchConfig(vo.SearchConfigParam{
		Search:   "accurate", // 精确搜索
		Group:    group,
		PageSize: 1000,
	})).(*model.ConfigPage)

	var data = make([]*watcher3.Response, len(cfgMap.PageItems))
	for i := range cfgMap.PageItems {
		var item = cfgMap.PageItems[i]
		data[i] = &watcher3.Response{
			Event: event.EventType_UPDATE,
			Key:   item.DataId,
			Value: strutil.ToBytes(item.Content),
		}
	}
	return data, nil
}

func (cm *nacosWatcher) GetCallback(ctx context.Context, key string, fn func(resp *watcher3.Response), opts ...watcher3.Opt) error {
	var dt, err = cm.Get(ctx, key, opts...)
	if err != nil {
		return err
	}

	for i := range dt {
		fn(dt[i])
	}
	return nil
}

func (cm *nacosWatcher) Watch(ctx context.Context, dataId string, opts ...watcher3.Opt) <-chan *watcher3.Response {
	resp := make(chan *watcher3.Response)
	configParams := vo.ConfigParam{
		DataId: dataId,
		Group:  cm.cfg.Group,
		OnChange: func(namespace, group, dataId, data string) {
			resp <- &watcher3.Response{
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

func (cm *nacosWatcher) WatchCallback(ctx context.Context, key string, fn func(resp *watcher3.Response), opts ...watcher3.Opt) {
	for resp := range cm.Watch(ctx, key, opts...) {
		fn(resp)
	}
}
