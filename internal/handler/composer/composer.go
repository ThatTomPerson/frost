package composer

import (
	"github.com/apex/log"
	"ttp.sh/frost/internal/project"
)

var DefaultHandler = new(handler)

func init() {
	project.AddHandler(DefaultHandler)
}

type handler struct{}

var _ project.PackageHandler = &handler{}

func (handler) Name() string {
	return "Composer"
}

func (handler) Detect(p *project.Project) bool {
	log.Debug("looking for composer.lock")

	return p.FileExists("composer.lock")
}

func (h *handler) GetLockFile(p *project.Project) Lock {

}

func (h *handler) GetModuleList(p *project.Project) (list []project.Module) {
	lock := h.GetLockFile(p)
	for _, pkg := range lock.Packages {
		mod := pkg.(project.Module)
		list = append(list, mod)
	}
	return
}
