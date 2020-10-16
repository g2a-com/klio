package registry

import (
	"runtime"

	"github.com/Masterminds/semver/v3"
	"github.com/g2a-com/klio/pkg/schema"
)

type Version schema.RegistryEntryVersion

func (ver *Version) Match(constraint string) bool {
	c, err1 := semver.NewConstraint(constraint)
	v, err2 := semver.NewVersion(ver.Number)
	return err1 == nil && err2 == nil && c.Check(v)
}

func (ver *Version) GreaterThan(ver2 Version) bool {
	v1, err1 := semver.NewVersion(ver.Number)
	v2, err2 := semver.NewVersion(ver2.Number)
	return err1 == nil && err2 == nil && v1.GreaterThan(v2)
}

func (ver *Version) IsCompatible() bool {
	return (ver.OS == runtime.GOOS || ver.OS == "*") && (ver.Arch == runtime.GOARCH || ver.Arch == "*")
}
