package manager

import (
	"sync"

	"github.com/apex/log"
	"github.com/vbauerster/mpb"
	"ttp.sh/frost/internal/project"
	"ttp.sh/frost/internal/semver"
)

type Handler interface {
	Name() string
	Enabled() bool
	// Install(*mpb.Progress)
	GetModuleList(project.Project) []Module
}

type Module interface {
	// Install from cache, if it is not in cache then Download first
	Install(project.Project)
	// Download the module into a package that can be installed to a location or cache
	Download() Package
	Name() string
	Version() *semver.Version
}

//
type Package interface {
	WriteTo(string) error
}

type Manager struct {
	root     string
	handlers []Handler
}

func (m *Manager) Add(h Handler) {
	m.handlers = append(m.handlers, h)
}

func New(root string) *Manager {
	return &Manager{root: root}
}

func (m *Manager) GetEnabledHandlers() (enabled []Handler) {
	for _, h := range m.handlers {
		log.Infof("%s is %v", h.Name(), h.Enabled())
		if h.Enabled() {
			enabled = append(enabled, h)
		}
	}

	return enabled
}

func (m *Manager) Run() error {
	handlers := m.GetEnabledHandlers()

	wg := new(sync.WaitGroup)
	wg.Add(len(handlers))

	pb := mpb.New(mpb.WithWaitGroup(wg))

	for _, h := range handlers {
		go func(h Handler) {
			h.GetModuleList()
			wg.Done()
		}(h)
	}

	pb.Wait()

	return nil
}
