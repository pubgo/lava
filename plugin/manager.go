package plugin

import (
	"encoding/json"
	"sync"

	"github.com/pubgo/x/stack"
	"github.com/pubgo/xerror"
)

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

func (m *manager) lock(fn func()) {
	m.Lock()
	defer m.Unlock()

	fn()
}

func (m *manager) All() map[string][]Plugin {
	pls := make(map[string][]Plugin)
	m.lock(func() {
		for k, v := range m.plugins {
			pls[k] = append(pls[k], v...)
		}
	})
	return pls
}

// Plugins lists the plugins
func (m *manager) Plugins(opts ...ManagerOpt) []Plugin {
	mOpts := managerOpts{Module: defaultModule}
	for _, o := range opts {
		o(&mOpts)
	}

	m.Lock()
	defer m.Unlock()

	if plugins, ok := m.plugins[mOpts.Module]; ok {
		return plugins
	}
	return nil
}

// Register registers a plugins
func (m *manager) Register(pg Plugin, opts ...ManagerOpt) {
	xerror.Assert(pg == nil, "plugin is nil")

	options := managerOpts{Module: defaultModule}
	for _, o := range opts {
		o(&options)
	}

	name := pg.String()
	xerror.Assert(name == "", "plugin name is null")

	m.Lock()
	defer m.Unlock()

	reg, ok := m.registered[options.Module]
	xerror.Assert(ok && reg[name], "Plugin [%s] Already Registered", name)

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
}

func (m *manager) isRegistered(plg Plugin, opts ...ManagerOpt) bool {
	options := managerOpts{Module: defaultModule}
	for _, o := range opts {
		o(&options)
	}

	m.Lock()
	defer m.Unlock()

	if _, ok := m.registered[options.Module]; !ok {
		return false
	}

	return m.registered[options.Module][stack.Func(plg)]
}
