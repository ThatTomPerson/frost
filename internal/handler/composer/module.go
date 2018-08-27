package composer

import (
	"archive/zip"
	"bytes"
	"context"
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
	"ttp.sh/frost/internal/project"
	"ttp.sh/frost/internal/semver"
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

// func (m *Module) Install(p *project.Project) {

// }

func (Module) GetHandler() project.PackageHandler {
	return DefaultHandler
}

func (h *Module) Install(pkg Module, root string) (Module, error) {
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

func (h *Module) InstallGit(ctx context.Context, root string, pkg Module) error {
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

func (h *Module) InstallZip(ctx context.Context, root string, pkg Module) error {
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
