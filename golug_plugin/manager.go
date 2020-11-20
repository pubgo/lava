package golug_plugin

import (
	"encoding/json"
	"sync"

	"github.com/pubgo/xerror"
	"github.com/pubgo/xerror/xerror_util"
)

// NewManager creates a new internal_plugin manager
func NewManager() Manager {
	return newManager()
}

func newManager() *manager {
	return &manager{
		plugins:    make(map[string][]Plugin),
		registered: make(map[string]map[string]bool),
	}
}

type manager struct {
	sync.Mutex
	plugins    map[string][]Plugin
	registered map[string]map[string]bool
}

func (m *manager) String() string {
	return string(xerror.PanicBytes(json.MarshalIndent(m.registered, "", "  ")))
}

func (m *manager) All() map[string][]Plugin {
	m.Lock()
	defer m.Unlock()

	pls := make(map[string][]Plugin)
	for k, v := range m.plugins {
		pls[k] = append(pls[k], v...)
	}
	return pls
}

func (m *manager) Plugins(opts ...ManagerOption) []Plugin {
	options := ManagerOptions{Module: defaultModule}
	for _, o := range opts {
		o(&options)
	}

	m.Lock()
	defer m.Unlock()

	if plugins, ok := m.plugins[options.Module]; ok {
		return plugins
	}
	return nil
}

func (m *manager) Register(pg Plugin, opts ...ManagerOption) error {
	if pg == nil {
		return xerror.New("plugin is nil")
	}

	options := ManagerOptions{Module: defaultModule}
	for _, o := range opts {
		o(&options)
	}

	name := pg.String()
	if name == "" {
		return xerror.New("plugin.name is null")
	}

	m.Lock()
	defer m.Unlock()

	if reg, ok := m.registered[options.Module]; ok && reg[name] {
		return xerror.Fmt("Plugin [%s] Already Registered", name)
	}

	if _, ok := m.registered[options.Module]; !ok {
		m.registered[options.Module] = map[string]bool{name: true}
	} else {
		m.registered[options.Module][name] = true
	}

	if _, ok := m.plugins[options.Module]; !ok {
		m.plugins[options.Module] = []Plugin{pg}
	} else {
		m.plugins[options.Module] = append(m.plugins[options.Module], pg)
	}
	return nil
	//return xerror.Wrap(golug_watcher.Watch(name, pg.Watch))
}

func (m *manager) isRegistered(plg Plugin, opts ...ManagerOption) bool {
	options := ManagerOptions{Module: defaultModule}
	for _, o := range opts {
		o(&options)
	}

	m.Lock()
	defer m.Unlock()

	if _, ok := m.registered[options.Module]; !ok {
		return false
	}

	return m.registered[options.Module][xerror_util.CallerWithFunc(plg)]
}
