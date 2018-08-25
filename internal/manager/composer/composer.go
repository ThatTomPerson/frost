package composer

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/apex/log"
	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
	"ttp.sh/frost/internal/manager"
	"ttp.sh/frost/internal/semver"
)

func New(root string) manager.Handler {
	return &handler{root}
}

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
	ContentHash string   `json:"content-hash"`
	Packages    []Module `json:"packages"`
	PackagesDev []Module `json:"packages-dev"`
}

type handler struct {
	root string
}

func (h *handler) Enabled() bool {
	_, err := os.Stat(path.Join(h.root, "composer.lock"))
	return err == nil
}

func (h *handler) Name() string {
	return "Composer"
}

func (h *handler) LoadLockFile() (lock *Lock, err error) {
	f, err := os.Open(path.Join(h.root, "composer.lock"))
	if err != nil {
		return nil, err
	}

	err = json.NewDecoder(f).Decode(&lock)
	if err != nil {
		return nil, err
	}

	return lock, nil
}

func (h *handler) download(pkg Module) Module {
	pkg, _ = h.Install(pkg, h.root)
	return pkg
}

func (h *handler) Install(pkg Module, root string) (Module, error) {
	var err error
	ctx := context.Background()
	switch pkg.Dist.Type {
	case "zip":
		err = h.InstallZip(ctx, root, pkg)
		pkg.InstallationSource = "dist"
	default:
		err = fmt.Errorf("No dist install")
	}

	if err != nil {
		switch pkg.Source.Type {
		case "git":
			err = h.InstallGit(ctx, root, pkg)
		}
		pkg.InstallationSource = "source"
	}

	if err != nil {
		return pkg, fmt.Errorf("Can't install package %s", pkg.Name)
	}

	v, _ := semver.NewVersion(pkg.Version)
	pkg.VersionNormalized = v.String()

	return pkg, nil
}

func (h *handler) InstallGit(ctx context.Context, root string, pkg Module) error {
	p := path.Join(root, "vendor", pkg.Name)
	_, err := os.Stat(p)
	if err != nil {
		clone := exec.Command("git", "clone", pkg.Source.URL, p)
		// clone.Stderr = os.Stderr
		// clone.Stdin = os.Stdin
		// clone.Stdout = os.Stdout
		clone.Run()
	}

	co := exec.Command("git", "checkout", pkg.Source.Reference)
	co.Dir = p
	// co.Stderr = os.Stderr
	// co.Stdin = os.Stdin
	// co.Stdout = os.Stdout
	co.Run()

	return nil
}

func (h *handler) InstallZip(ctx context.Context, root string, pkg Module) error {
	r, err := http.Get(pkg.Dist.URL)
	if err != nil {
		return fmt.Errorf("%s: %s: %v", h.Name(), pkg.Name, err)
	}

	defer r.Body.Close()

	// ReadAll reads from readCloser until EOF and returns the data as a []byte
	b, err := ioutil.ReadAll(r.Body) // The readCloser is the one from the zip-package
	if err != nil {
		log.Fatalf("%v", err)
	}

	// bytes.Reader implements io.Reader, io.ReaderAt, etc. All you need!
	readerAt := bytes.NewReader(b)

	z, err := zip.NewReader(readerAt, int64(len(b)))
	if err != nil {
		return fmt.Errorf("%s: %s: %v", h.Name(), pkg.Name, err)
	}

	for _, f := range z.File {
		parts := strings.Split(f.Name, "/")
		name := strings.Join(parts[1:], "/")
		p := path.Join(root, "vendor", pkg.Name, name)

		if f.FileInfo().IsDir() {
			err := os.MkdirAll(p, os.ModePerm)
			if err != nil {
				log.Errorf("%v", err)
			}
		} else {
			out, err := os.Create(p)
			if err != nil {
				log.Errorf("%v", err)
			}
			reader, err := f.Open()
			if err != nil {
				log.Errorf("%v", err)
			}
			_, err = io.Copy(out, reader)
			if err != nil {
				log.Errorf("%v", err)
			}
			err = reader.Close()
			if err != nil {
				log.Errorf("%v", err)
			}
		}
	}

	return nil
}

func (h *handler) downloadWorker(c chan Module, installed chan Module, bar *mpb.Bar, wg *sync.WaitGroup) {
	for pkg := range c {
		start := time.Now()
		installed <- h.download(pkg)
		bar.IncrBy(1, time.Since(start))
	}

	wg.Done()
}

func (h *handler) downloadWorkers(bar *mpb.Bar) (chan Module, chan Module, *sync.WaitGroup) {
	workers := runtime.NumCPU() * 4

	wg := new(sync.WaitGroup)
	wg.Add(workers)
	c := make(chan Module, workers*2)
	installed := make(chan Module, workers*2)

	for i := 0; i < workers; i++ {
		go h.downloadWorker(c, installed, bar, wg)
	}

	return c, installed, wg
}

func (h *handler) Run(pb *mpb.Progress) {
	lock, err := h.LoadLockFile()
	if err != nil {
		return
	}

	total := len(lock.Packages) + len(lock.PackagesDev)

	bar := pb.AddBar(int64(total),
		mpb.BarRemoveOnComplete(),
		mpb.PrependDecorators(
			decor.Name(h.Name(), decor.WC{W: len(h.Name()) + 1, C: decor.DidentRight}),
			decor.Name("Downloading", decor.WCSyncSpaceR),
			decor.CountersNoUnit("%d / %d", decor.WCSyncWidth),
		),
		mpb.AppendDecorators(decor.Percentage(decor.WC{W: 5})))

	c, installed, wg := h.downloadWorkers(bar)
	done := make(chan bool)
	go func() {
		var pkgs []Module
		for pkg := range installed {
			pkgs = append(pkgs, pkg)
		}

		os.MkdirAll(path.Join(h.root, "vendor/composer"), os.ModePerm)
		f, _ := os.Create(path.Join(h.root, "vendor/composer/installed.json"))

		json.NewEncoder(f).Encode(pkgs)
		f.Close()

		done <- true
	}()

	for _, pkg := range lock.Packages {
		c <- pkg
	}
	for _, pkg := range lock.PackagesDev {
		c <- pkg
	}

	close(c)
	wg.Wait()
	close(installed)

	<-done

	cmd := exec.Command("composer", "dump-autoload")
	cmd.Run()
}
