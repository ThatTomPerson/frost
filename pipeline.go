package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	// "archive/zip"
	"github.com/klauspost/compress/zip"

	"github.com/sirupsen/logrus"
)

type ClassMap struct {
	// "ClassName": "path from base"
	Classes map[string]string
	// "hash": "path from base"
	Files map[string]string
}

func NewClassMap() *ClassMap {
	return &ClassMap{
		Classes: make(map[string]string),
		Files:   make(map[string]string),
	}
}

func (cm *ClassMap) AddClass(m Module, path string, body []byte) {
	if filepath.Ext(path) != ".php" {
		return
	}

	for prefix, root := range m.Autoload.Psr4 {
		// spew.Dump(path, root)
		if strings.Index(path, root) == 0 {

			classPath := prefix + strings.Replace(path[len(root):len(path)-4], "/", "\\", -1)
			cm.Classes[classPath] = "vendor/" + m.Name + "/" + path
		}
	}
}

func (cm *ClassMap) Merge(_cm *ClassMap) {
	for classPath, path := range _cm.Classes {
		cm.Classes[classPath] = path
	}
}

func download(m Module) *ClassMap {
	logrus.Infof("installing %s", m.Name)
	res, err := http.Get(m.Dist.URL)
	if err != nil {
		logrus.Warnf("[%s] %v", m.Name, err)
	}
	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logrus.Warnf("[%s] %v", m.Name, err)
	}

	buf := bytes.NewReader(b)

	r, err := zip.NewReader(buf, buf.Size())
	if err != nil {
		logrus.Warnf("[%s] %v", m.Name, err)
	}

	cm := NewClassMap()

	for _, f := range r.File {

		ix := strings.Index(f.Name, "/")
		path := f.Name[ix+1:]
		fullpath := fmt.Sprintf("./vendor/%s/%s", m.Name, path)
		if f.Mode().IsDir() {
			os.MkdirAll(fullpath, os.ModePerm)
		} else {
			fc, _ := f.Open()
			if m.CouldBeClass(path) {
				b, _ := ioutil.ReadAll(fc)
				cm.AddClass(m, path, b)
				ioutil.WriteFile(fullpath, b, os.ModePerm)
			} else {
				o, _ := os.Create(fullpath)
				io.Copy(o, fc)
				o.Close()
			}
			fc.Close()

		}
	}

	return cm
}

// open
//  * download + classmap
// collect
// generate
