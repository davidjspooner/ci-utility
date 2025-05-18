package semantic

import (
	"fmt"
	"regexp"
	"strconv"
)

type Version struct {
	Major int
	Minor int
	Patch int
}

func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}
func (v Version) Increment(bump string) (Version, error) {
	switch bump {
	case "major":
		return Version{v.Major + 1, 0, 0}, nil
	case "minor":
		return Version{v.Major, v.Minor + 1, 0}, nil
	case "patch":
		return Version{v.Major, v.Minor, v.Patch + 1}, nil
	default:
		return Version{}, fmt.Errorf("unknown version bump type: %q", bump)
	}
}
func (v Version) IsValid() bool {
	return v.Major >= 0 && v.Minor >= 0 && v.Patch >= 0
}
func (v Version) Compare(other Version) int {
	if v.Major != other.Major {
		return v.Major - other.Major
	}
	if v.Minor != other.Minor {
		return v.Minor - other.Minor
	}
	return v.Patch - other.Patch
}
func (v Version) IsGreaterThan(other Version) bool {
	return v.Compare(other) > 0
}
func (v Version) IsLessThan(other Version) bool {
	return v.Compare(other) < 0
}
func (v Version) IsEqual(other Version) bool {
	return v.Compare(other) == 0
}
func (v Version) IsGreaterThanOrEqual(other Version) bool {
	return v.Compare(other) >= 0
}
func (v Version) IsLessThanOrEqual(other Version) bool {
	return v.Compare(other) <= 0
}
func (v Version) IsZero() bool {
	return v.Major == 0 && v.Minor == 0 && v.Patch == 0
}
func (v Version) IsEmpty() bool {
	return v.Major == 0 && v.Minor == 0 && v.Patch == 0
}
func (v Version) IsNotEmpty() bool {
	return !v.IsEmpty()
}

var versionFmt = regexp.MustCompile(`(.*)(\d+)\.(\d+)\.(\d+)(.*)`)

func ExtractVersionFromTag(tag string) (string, string, Version, error) {
	matches := versionFmt.FindStringSubmatch(tag)
	if len(matches) != 6 {
		return "", "", Version{}, fmt.Errorf("invalid version tag format: %s", tag)
	}
	v := Version{}
	var err error
	v.Major, err = strconv.Atoi(matches[2])
	if err != nil {
		return "", "", v, fmt.Errorf("error converting major version: %v", err)
	}
	v.Minor, err = strconv.Atoi(matches[3])
	if err != nil {
		return "", "", v, fmt.Errorf("error converting minor version: %v", err)
	}
	v.Patch, err = strconv.Atoi(matches[4])
	if err != nil {
		return "", "", v, fmt.Errorf("error converting patch version: %v", err)
	}
	return matches[1], matches[5], v, nil
}
