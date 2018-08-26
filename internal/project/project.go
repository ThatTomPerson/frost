package project

import (
	"os"
	"path"
	"runtime"
	"sync"

	"github.com/apex/log"
)

var handlers []PackageHandler

func AddHandler(h PackageHandler) {
	handlers = append(handlers, h)
}

type Project struct {
	root     string
	Handlers []PackageHandler
}

func (p *Project) GetFile(f string) *os.File {
	file, _ := os.Open(path.Join(p.root, f))
	return file
}

func (p *Project) detectPackageHandlers() {
	log.Infof("Detected package managers")
	for _, h := range handlers {
		if h.Detect(p) {
			log.Infof("  %s", h.Name())

			p.Handlers = append(p.Handlers, h)
		}
	}
}

func (p *Project) installWorker(c chan Module, wg *sync.WaitGroup) {
	for mod := range c {
		mod.Install(p)
		mod.GetHandler()
	}

	wg.Done()
}

func (p *Project) installWorkers() (chan Module, *sync.WaitGroup) {
	workers := runtime.NumCPU() * 4

	wg := new(sync.WaitGroup)
	wg.Add(workers)

	c := make(chan Module, workers*2)

	for i := 0; i < workers; i++ {
		go p.installWorker(c, wg)
	}

	return c, wg
}

func (p *Project) Install() {

	list := p.GetModuleList()

	log.Infof("found %d modules", len(list))

	c, wg := p.installWorkers()

	for _, mod := range list {
		c <- mod
	}

	close(c)
	wg.Wait()
}

func (p *Project) GetModuleList() (modules []Module) {
	for _, h := range p.Handlers {
		modules = append(modules, h.GetModuleList(p)...)
	}

	return
}

func (p *Project) FileExists(f string) bool {
	_, err := os.Stat(path.Join(p.root, f))
	return err == nil
}

type PackageHandler interface {
	Name() string
	Detect(*Project) bool
	GetModuleList(*Project) []Module
}

func New(root string) *Project {
	p := &Project{root: root}

	p.detectPackageHandlers()

	return p
}
