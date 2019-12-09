package registry

import (
	"fmt"
	"regexp"

	"github.com/Masterminds/semver"
)

const versionRegexp = `^v(?P<Version>\d+\.\d+\.\d(-\w.+)?(\+\w.+)?)-(?P<OS>\w+)-(?P<Arch>\w+)$`

// CommandVersion stores information about command version
type CommandVersion struct {
	Version *semver.Version
	OS      string
	Arch    string
}

func (version *CommandVersion) String() string {
	return fmt.Sprintf("v%s-%s-%s", version.Version, version.OS, version.Arch)
}

// NewCommandVersion parses a given version string and returns an instance of
// CommandVersion or an error if unable to parse the version.
func NewCommandVersion(versionString string) (*CommandVersion, error) {
	groups := matchAgainstGroups(versionRegexp, versionString)
	if len(groups) == 0 {
		return nil, fmt.Errorf("invalid version string: %s", versionString)
	}

	ver, err := semver.NewVersion(groups["Version"])
	if err != nil {
		return nil, fmt.Errorf("invalid version string: %s", versionString)
	}

	version := &CommandVersion{
		Version: ver,
		Arch:    groups["Arch"],
		OS:      groups["OS"],
	}

	return version, nil
}

// CommandVersionSet represents list of versions of single command
type CommandVersionSet []CommandVersion

// MatchVersion returns greatest version matching to the specified parameters
func (versionSet *CommandVersionSet) MatchVersion(versionConstraints *semver.Constraints, os string, arch string) (*CommandVersion, bool) {
	var bestMatch CommandVersion

	for _, version := range *versionSet {
		if version.Version == nil {
			continue
		}
		if versionConstraints.Check(version.Version) == false {
			continue
		}
		if bestMatch.Version != nil {
			if version.OS != "" && os != "" && version.OS != os {
				continue
			}
			if version.Arch != "" && arch != "" && version.Arch != arch {
				continue
			}
			if version.Version.LessThan(bestMatch.Version) {
				continue
			}
			if version.Arch == "" && bestMatch.Arch != "" {
				continue
			}
			if version.OS == "" && bestMatch.Arch != "" {
				continue
			}
		}
		bestMatch = version
	}

	return &bestMatch, bestMatch.Version != nil
}

func (versionSet *CommandVersionSet) String() string {
	var result string
	for _, version := range *versionSet {
		if result != "" {
			result += ", "
		}
		result += version.String()
	}
	return result
}

func matchAgainstGroups(regularExpression string, url string) map[string]string {
	var compRegEx = regexp.MustCompile(regularExpression)
	match := compRegEx.FindStringSubmatch(url)

	paramsMap := make(map[string]string)
	for i, name := range compRegEx.SubexpNames() {
		if i > 0 && i <= len(match) {
			paramsMap[name] = match[i]
		}
	}

	return paramsMap
}
