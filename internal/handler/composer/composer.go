package composer

import (
	"encoding/json"

	"github.com/apex/log"
	"ttp.sh/frost/internal/project"
)

var DefaultHandler = new(handler)

const lockFile = "composer.lock"

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

	return p.FileExists(lockFile)
}

func (h *handler) GetLockFile(p *project.Project) (lock Lock) {
	f := p.GetFile(lockFile)
	json.NewDecoder(f).Decode(&lock)
	return
}

func (h *handler) GetModuleList(p *project.Project) (list []project.Module) {
	lock := h.GetLockFile(p)
	for _, pkg := range lock.Packages {
		p := pkg
		mod := project.Module(&p)
		list = append(list, mod)
	}
	return
}
