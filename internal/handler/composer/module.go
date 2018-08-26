package composer

import (
	"time"

	"ttp.sh/frost/internal/project"
)

type Module struct {
	Name              string `json:"name"`
	Version           string `json:"version"`
	VersionNormalized string `json:"version_normalized"`
	Source            struct {
		Type      string `json:"type"`
		URL       string `json:"url"`
		Reference string `json:"reference"`
	} `json:"source"`
	Dist struct {
		Type      string `json:"type"`
		URL       string `json:"url"`
		Reference string `json:"reference"`
		Shasum    string `json:"shasum"`
	} `json:"dist"`
	Require            map[string]string `json:"require"`
	Conflict           map[string]string `json:"conflict"`
	Replace            map[string]string `json:"replace"`
	RequireDev         map[string]string `json:"require-dev"`
	Suggest            map[string]string `json:"suggest"`
	Time               time.Time         `json:"time"`
	Type               string            `json:"type"`
	Extra              interface{}       `json:"extra"`
	InstallationSource string            `json:"installation-source"`
	Autoload           interface{}       `json:"autoload"`
	NotificationURL    string            `json:"notification-url"`
	License            []string          `json:"license"`
	Authors            []struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Homepage string `json:"homepage"`
	} `json:"authors"`
	Description string   `json:"description"`
	Keywords    []string `json:"keywords"`

	installed bool
}

type Lock struct {
	ContentHash string   `json:"content-hash"`
	Packages    []Module `json:"packages"`
	PackagesDev []Module `json:"packages-dev"`
}

func (m *Module) Install(p *project.Project) {
	
}


func (Module) GetHandler() project.PackageHandler {
	return DefaultHandler
}
