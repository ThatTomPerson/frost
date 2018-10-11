package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
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
}

type Lock struct {
	Packages    []Module `json:"packages"`
	PackagesDev []Module `json:"packages-dev"`
}

func install(p Module) {
	logrus.Infof("installing %s", p.Name)
	res, err := http.Get(p.Dist.URL)
	if err != nil {
		logrus.Warnf("[%s] %v", p.Name, err)
	}
	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logrus.Warnf("[%s] %v", p.Name, err)
	}

	buf := bytes.NewReader(b)

	r, err := zip.NewReader(buf, buf.Size())
	if err != nil {
		logrus.Warnf("[%s] %v", p.Name, err)
	}

	for _, f := range r.File {
		// fmt.Printf("Contents of %s:\n", f.Name)
		ix := strings.Index(f.Name, "/")
		name := fmt.Sprintf("./vendor/%s/%s", p.Name, f.Name[ix:])
		if f.Mode().IsDir() {
			os.MkdirAll(name, os.ModePerm)
		} else {
			o, _ := os.Create(name)
			fc, _ := f.Open()
			io.Copy(o, fc)
			fc.Close()
			o.Close()
		}
	}
}

func installer(c chan Module, wg *sync.WaitGroup) {
	for p := range c {
		install(p)
	}

	wg.Done()
}

func main() {

	logrus.Info("cleaning vendor dir")
	os.RemoveAll("vendor")

	logrus.Info("opening composer.lock")
	f, err := os.Open("composer.lock")
	if err != nil {
		logrus.Fatal(err)
	}

	var l Lock
	json.NewDecoder(f).Decode(&l)

	logrus.Info("installing packages")

	c := make(chan Module, 50)
	wg := new(sync.WaitGroup)
	wg.Add(20)

	for i := 0; i < 20; i++ {
		go installer(c, wg)
	}

	for _, p := range l.PackagesDev {
		c <- p
	}
	for _, p := range l.Packages {
		c <- p
	}
	close(c)
	wg.Wait()
}
