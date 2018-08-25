package semver

import (
	"fmt"
	"strings"

	"github.com/coreos/go-semver/semver"
)

type Version struct {
	*semver.Version

	version string
}

func NewVersion(version string) (*Version, error) {
	if version[0] == 'v' {
		version = version[1:]
	}
	if len(version) > 3 && version[:3] == "dev" {
		version = "9999999-dev"

		return &Version{version: version}, nil
	}

	parts := strings.Split(version, ".")
	for i := len(parts); i < 3; i++ {
		parts = append(parts, "0")
	}
	version = strings.Join(parts, ".")

	v, err := semver.NewVersion(version)
	return &Version{version: version, Version: v}, err

}

func (v *Version) String() string {
	if !v.IsGitRef() {
		return fmt.Sprintf("%s.%d", v.Version.String(), 0)
	}

	return v.version
}

func (v *Version) IsGitRef() bool {
	return v.version == "9999999-dev"
}
