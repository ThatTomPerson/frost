package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
)

func test1() {
	// b, err := os.Open("composer.json")
	// if err != nil {
	// 	log.Fatal("couldn't find file")
	// }

	// var r RootPackage
	// json.NewDecoder(b).Decode(&r)
	// log.Info(r.Name)

	// log.Info("Resolving Dependancies")
	// _, _ = resolveDependancies(r)

	// for name, _ := range r.Require {
	// 	download(name)
	// }
}

func resolveDependanciesWorker(deps chan string, wg *sync.WaitGroup) {
	for dep := range deps {
		if dep == "php" {
			continue
		}

		// get(dep)
		// log.Info(dep)
	}

	wg.Done()
}

func resolveDependancies(r RootPackage) ([]string, error) {
	workers := runtime.NumCPU() * 4

	wg := &sync.WaitGroup{}
	depChan := make(chan string, workers*4)
	for i := 0; i < workers; i++ {
		go resolveDependanciesWorker(depChan, wg)
	}

	for dep, _ := range r.Require {
		depChan <- dep
	}

	for dep, _ := range r.RequireDev {
		depChan <- dep
	}

	wg.Wait()

	return []string{}, nil
}

func download(name string) {
	r, err := http.Get(fmt.Sprintf("https://repo.packagist.org/p/%s.json", name))
	if err != nil {
		return
	}
	defer r.Body.Close()
	f, err := os.Create(fmt.Sprintf("./cache/%s.json", strings.Replace(name, "/", "-", -1)))
	if err != nil {
		return
	}

	io.Copy(f, r.Body)
	f.Close()
}
