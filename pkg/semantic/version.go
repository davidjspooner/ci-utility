package semantic

import (
	"fmt"
	"regexp"
	"strconv"
)

// Version represents a semantic version (major.minor.patch).
type Version struct {
	Major int
	Minor int
	Patch int
}

// String returns the version as a string in the form "major.minor.patch".
func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// Increment returns a new Version incremented by the specified bump type ("major", "minor", or "patch").
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

// IsValid returns true if the version is valid (all parts are non-negative).
func (v Version) IsValid() bool {
	return v.Major >= 0 && v.Minor >= 0 && v.Patch >= 0
}

// Compare compares two versions. Returns positive if v > other, negative if v < other, zero if equal.
func (v Version) Compare(other Version) int {
	if v.Major != other.Major {
		return v.Major - other.Major
	}
	if v.Minor != other.Minor {
		return v.Minor - other.Minor
	}
	return v.Patch - other.Patch
}

// IsGreaterThan returns true if v is greater than other.
func (v Version) IsGreaterThan(other Version) bool {
	return v.Compare(other) > 0
}

// IsLessThan returns true if v is less than other.
func (v Version) IsLessThan(other Version) bool {
	return v.Compare(other) < 0
}

// IsEqual returns true if v is equal to other.
func (v Version) IsEqual(other Version) bool {
	return v.Compare(other) == 0
}

// IsGreaterThanOrEqual returns true if v is greater than or equal to other.
func (v Version) IsGreaterThanOrEqual(other Version) bool {
	return v.Compare(other) >= 0
}

// IsLessThanOrEqual returns true if v is less than or equal to other.
func (v Version) IsLessThanOrEqual(other Version) bool {
	return v.Compare(other) <= 0
}

// IsZero returns true if the version is 0.0.0.
func (v Version) IsZero() bool {
	return v.Major == 0 && v.Minor == 0 && v.Patch == 0
}

// IsEmpty returns true if the version is 0.0.0.
func (v Version) IsEmpty() bool {
	return v.Major == 0 && v.Minor == 0 && v.Patch == 0
}

// IsNotEmpty returns true if the version is not 0.0.0.
func (v Version) IsNotEmpty() bool {
	return !v.IsEmpty()
}

var versionFmt = regexp.MustCompile(`(.*)(\d+)\.(\d+)\.(\d+)(.*)`)

// ExtractVersionFromTag extracts a semantic version from a tag string.
func ExtractVersionFromTag(tag string) (string, string, Version, error) {
	// Match the tag against the version format.
	matches := versionFmt.FindStringSubmatch(tag)
	if len(matches) != 6 {
		return "", "", Version{}, fmt.Errorf("invalid version tag format: %s", tag)
	}
	v := Version{}
	var err error

	// Convert the major, minor, and patch parts to integers.
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
	//return the prefix, suffix, version struct, and nil error
	return matches[1], matches[5], v, nil
}
