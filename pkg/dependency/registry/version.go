package registry

import (
	"github.com/Masterminds/semver/v3"
)

type Version string

func (ver Version) Match(constraint string) bool {
	c, err1 := semver.NewConstraint(constraint)
	v, err2 := semver.NewVersion(string(ver))
	return err1 == nil && err2 == nil && c.Check(v)
}

func (ver Version) GreaterThan(ver2 Version) bool {
	v1, err1 := semver.NewVersion(string(ver))
	v2, err2 := semver.NewVersion(string(ver2))
	return err1 == nil && err2 == nil && v1.GreaterThan(v2)
}
