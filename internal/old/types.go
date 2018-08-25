package main

import (
	"time"

	"github.com/masterminds/semver"
)

type Dependancies map[string]string

type Dependancy struct{}

type Autoload struct {
	Classmap []string               `json:"classmap"`
	Psr4     map[string]interface{} `json:"psr-4"`
}

type RootPackage struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Keywords    []string     `json:"keywords"`
	License     string       `json:"license"`
	Type        string       `json:"type"`
	Require     Dependancies `json:"require"`
	RequireDev  Dependancies `json:"require-dev"`
	Autoload    Autoload     `json:"autoload"`
	AutoloadDev Autoload     `json:"autoload-dev"`
	Scripts     struct {
		PostRootPackageInstall []string `json:"post-root-package-install"`
		PostCreateProjectCmd   []string `json:"post-create-project-cmd"`
		PostAutoloadDump       []string `json:"post-autoload-dump"`
	} `json:"scripts"`
	Config struct {
		PreferredInstall   string `json:"preferred-install"`
		SortPackages       bool   `json:"sort-packages"`
		OptimizeAutoloader bool   `json:"optimize-autoloader"`
	} `json:"config"`
	MinimumStability string `json:"minimum-stability"`
	PreferStable     bool   `json:"prefer-stable"`
}

type ApiPackage struct {
	Name              string         `json:"name"`
	Description       string         `json:"description"`
	Keywords          []string       `json:"keywords"`
	Homepage          string         `json:"homepage"`
	Version           semver.Version `json:"version"`
	VersionNormalized semver.Version `json:"version_normalized"`
	License           []string       `json:"license"`
	Authors           []struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Homepage string `json:"homepage"`
	} `json:"authors"`
	Source struct {
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
	Type     string       `json:"type"`
	Time     time.Time    `json:"time"`
	Autoload Autoload     `json:"autoload"`
	Require  Dependancies `json:"require"`
	UID      int          `json:"uid"`
}

type ApiResponse struct {
	Packages map[string]map[string]ApiPackage `json:"packages"`
}
