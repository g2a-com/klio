package registry

import (
	"fmt"

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

func getExactMatch(version Version) (string, error) {
	_, err := semver.NewVersion(string(version))
	if err != nil && version != "*" {
		return "", err
	}
	return string(version), nil
}

func getMinorAndPatchConstraints(version Version) (string, error) {
	v, err := semver.NewVersion(string(version))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("^%s,>%s", v, v), nil
}

func getMajorConstraints(version Version) (string, error) {
	v, err := semver.NewVersion(string(version))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d.x", v.IncMajor().Major()), nil
}
