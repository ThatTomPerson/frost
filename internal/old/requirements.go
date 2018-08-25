package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"sync"

	"net/http"
	"os"
	"runtime"

	"github.com/apex/log"
)

func New(lock string) (*Meta, error) {
	pkg, lck, err := open(lock)
	if err != nil {
		return nil, err
	}

	return &Meta{
		pkg: pkg,
		lck: lck,
	}, nil
}

func open(lock string) (pkg *RootPackage, lck *Lock, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}

	err = json.NewDecoder(f).Decode(&pkg)
	f.Close()

	if err != nil {
		return nil, nil, err
		// log.Fatalf("err: %v", err)
	}

	f, err = os.Open(lock)
	if err != nil {
		log.Infof("No Lockfile: %v", err)
		return nil, nil, err
	}
	err = json.NewDecoder(f).Decode(&lck)
	f.Close()

	if err != nil {
		return nil, nil, err
	}

	return pkg, lck, err
}

type Meta struct {
	pkg          *RootPackage
	lck          *Lock
	requirements []Requirement
}

type Requirement struct {
	Package string
	Version string
}

func (m *Meta) AddRequirement(req Requirement) {
	m.requirements = append(m.requirements, req)
}

func (m *Meta) ParseRequirements() {

}

func (m *Meta) downloader(pkgs chan LockPackage, wg *sync.WaitGroup) {
	for pkg := range pkgs {
		log.Infof("Downloading %s", pkg.Name)

		r, err := http.Get(pkg.Dist.URL)
		if err != nil {
			log.Errorf("%v", err)
		}

		defer r.Body.Close()

		switch pkg.Dist.Type {
		case "zip":
			// ReadAll reads from readCloser until EOF and returns the data as a []byte
			b, err := ioutil.ReadAll(r.Body) // The readCloser is the one from the zip-package
			if err != nil {
				log.Fatalf("%v", err)
			}

			// bytes.Reader implements io.Reader, io.ReaderAt, etc. All you need!
			readerAt := bytes.NewReader(b)

			z, err := zip.NewReader(readerAt, int64(len(b)))
			if err != nil {
				log.Fatalf("%v", err)
			}

			for _, f := range z.File {

				parts := strings.Split(f.Name, "/")
				name := strings.Join(parts[1:], "/")

				if f.FileInfo().IsDir() {
					dirPath := fmt.Sprintf("./vendor/%s/%s", pkg.Name, name)
					err := os.MkdirAll(dirPath, os.ModePerm)
					if err != nil {
						log.Fatalf("%v", err)
					}
				} else {
					fPath := fmt.Sprintf("./vendor/%s/%s", pkg.Name, name)

					out, err := os.Create(fPath)
					if err != nil {
						log.Fatalf("%v", err)
					}
					reader, err := f.Open()
					if err != nil {
						log.Fatalf("%v", err)
					}
					_, err = io.Copy(out, reader)
					if err != nil {
						log.Fatalf("%v", err)
					}
				}

			}
		}

	}

	wg.Done()
}

func (m *Meta) Install() error {
	log.Info("Starting Install")

	workers := runtime.NumCPU() * 4

	pkgs := make(chan LockPackage)

	wg := &sync.WaitGroup{}
	wg.Add(workers)

	log.Info("Starting download workers")
	for i := 0; i < workers; i++ {
		go m.downloader(pkgs, wg)
	}

	for _, pkg := range m.lck.Packages {
		// log.Infof("Installing %s", pkg.Name)
		pkgs <- pkg
	}

	for _, pkg := range m.lck.PackagesDev {
		// log.Infof("Installing %s", pkg.Name)
		pkgs <- pkg
	}

	close(pkgs)

	wg.Wait()

	return nil
}
