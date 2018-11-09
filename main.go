package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
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
	Autoload           struct {
		Psr4     map[string]string `json:"psr-4"`
		Psr0     map[string]string `json:"psr-0"`
		Classmap []string          `json:"classmap"`
	} `json:"autoload"`
	NotificationURL string   `json:"notification-url"`
	License         []string `json:"license"`
	Authors         []struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Homepage string `json:"homepage"`
	} `json:"authors"`
	Description string   `json:"description"`
	Keywords    []string `json:"keywords"`
}

func (m *Module) CouldBeClass(path string) bool {
	ext := filepath.Ext(path)
	if ext != ".php" && ext != ".inc" && ext != ".hh" {
		return false // wrong ext
	}

	for _, root := range m.Autoload.Psr4 {
		if strings.Index(path, root) <= 1 {
			return true
		}
	}
	return false
}

type Lock struct {
	Packages    []Module `json:"packages"`
	PackagesDev []Module `json:"packages-dev"`
}

func installer(c chan Module, wg *sync.WaitGroup, cm *ClassMap, i map[string]Module) {
	for p := range c {
		if a, ok := i[p.Name]; ok {
			if a.Version == p.Version {
				continue
			}
		}
		cm.Merge(download(p))
		i[p.Name] = p
	}

	wg.Done()
}

func main() {

	// logrus.Info("cleaning vendor dir")
	// os.RemoveAll("vendor")

	logrus.Info("opening composer.lock")
	f, err := os.Open("composer.lock")
	if err != nil {
		logrus.Fatal(err)
	}

	var l Lock
	json.NewDecoder(f).Decode(&l)

	logrus.Info("opening installed.json")
	i := make(map[string]Module)
	f, err = os.Open("vendor/composer/installed.json")
	if err == nil {
		var _i []Module
		json.NewDecoder(f).Decode(_i)
		f.Close()
		for _, p := range _i {
			i[p.Name] = p
		}
	}

	logrus.Info("installing packages")

	c := make(chan Module, 50)
	wg := new(sync.WaitGroup)
	wg.Add(20)

	cm := NewClassMap()
	for j := 0; j < 20; j++ {
		go installer(c, wg, cm, i)
	}

	for _, p := range l.PackagesDev {
		c <- p
	}
	for _, p := range l.Packages {
		c <- p
	}
	close(c)
	wg.Wait()

	os.Mkdir("vendor/composer", os.ModePerm)
	f, err = os.Create("vendor/composer/installed.json")
	v := []Module{}
	for _, k := range i {
		v = append(v, k)
	}

	json.NewEncoder(f).Encode(v)
	f.Close()

}
