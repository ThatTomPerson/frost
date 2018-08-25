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
	"strings"
	"time"

	"github.com/apex/log"
	"ttp.sh/frost/internal/manager"
)

type Package struct {
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
	RequireDev         map[string]string `json:"require-dev"`
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
	ContentHash string    `json:"content-hash"`
	Packages    []Package `json:"packages"`
	PackagesDev []Package `json:"packages-dev"`
}

type handler struct {
}

func init() {
	manager.AddHandler(&handler{})
}

func (h *handler) Name() string {
	return "composer"
}

func (h *handler) Detect(root string) bool {
	_, err := os.Stat(path.Join(root, "composer.lock"))
	return err == nil
}

func extract(pkg Package) error {
	return nil
}

func (h *handler) Install(pkg Package, root string) manager.Job {
	return func(ctx context.Context) (err error) {
		log.Infof("Downloading %s", pkg.Name)

		switch pkg.Dist.Type {
		case "zip":
			err = h.InstallZip(ctx, root, pkg)
		default:
			err = fmt.Errorf("No dist install")
		}

		if err != nil {
			switch pkg.Source.Type {
			case "git":
				err = h.InstallGit(ctx, root, pkg)
			}
		}

		if err != nil {
			return fmt.Errorf("Can't install package %s", pkg.Name)
		}

		return nil
	}
}

func (h *handler) InstallGit(ctx context.Context, root string, pkg Package) error {
	p := path.Join(root, "vendor", pkg.Name)

	clone := exec.Command("git", "clone", pkg.Source.URL, p)
	clone.Stderr = os.Stderr
	clone.Stdin = os.Stdin
	clone.Stdout = os.Stdout
	clone.Run()

	co := exec.Command("git", "checkout", pkg.Source.Reference)
	co.Dir = p
	co.Stderr = os.Stderr
	co.Stdin = os.Stdin
	co.Stdout = os.Stdout
	co.Run()

	return nil
}

func (h *handler) InstallZip(ctx context.Context, root string, pkg Package) error {
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

func (h *handler) Exec(ctx context.Context, root string, jobs chan manager.Job) (err error) {
	// parse file
	f, err := os.Open("composer.lock")
	if err != nil {
		return err
	}

	var lck Lock
	err = json.NewDecoder(f).Decode(&lck)
	if err != nil {
		return err
	}
	f.Close()

	for _, dep := range lck.Packages {
		jobs <- h.Install(dep, root)
	}
	for _, dep := range lck.PackagesDev {
		jobs <- h.Install(dep, root)
	}

	var pkgs []Package
	for _, pkg := range lck.Packages {
		pkg.VersionNormalized = strings.Replace(fmt.Sprintf("%s.%d", pkg.Version, 0), "v", "", -1)
		pkgs = append(pkgs, pkg)
	}
	for _, pkg := range lck.PackagesDev {
		pkg.VersionNormalized = fmt.Sprintf("%s.%d", pkg.Version, 0)
		pkgs = append(pkgs, pkg)
	}

	os.MkdirAll("vendor/composer", os.ModePerm)
	f, _ = os.Create("vendor/composer/installed.json")

	json.NewEncoder(f).Encode(pkgs)
	f.Close()

	return nil
}

func (h *handler) PostExec(ctx context.Context, root string) error {

	log.Info("composer install")
	cmd := exec.Command("composer", "install")
	cmd.Dir = root
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	return cmd.Run()
	// return nil
}
