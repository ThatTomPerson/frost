package composer

import "ttp.sh/frost/internal/project"

type module struct {
}

func (m *module) Install(p *project.Project) {

}

func (module) GetHandler() project.PackageHandler {
	return DefaultHandler
}
