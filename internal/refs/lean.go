package refs

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"stewdio/internal/utils"
)

const ObjectTarName = "lean.tar.gz"

type Version struct {
	Major int `json:"major"`
	Minor int `json:"minor"`
}

func (v Version) String() string {
	return fmt.Sprintf("%d.%d", v.Major, v.Minor)
}

type Diff struct {
	File string `json:"file"`
	Type string `json:"type"`
}

func WriteVersion(path string, version Version) {
	versionStr := fmt.Sprintf("%d.%d", version.Major, version.Minor)

	versionPath := filepath.Join(path, ".stew", "version")

	if err := os.WriteFile(versionPath, []byte(versionStr), 0o644); err != nil {
		panic(err)
	}
}

func IsStewRepo(path string) bool {
	return utils.PathExists(filepath.Join(path, ".stew"))
}

func ReadVersion(path string) Version {
	versionPath := filepath.Join(path, ".stew", "version")

	data, err := os.ReadFile(versionPath)
	if err != nil {
		panic(err)
	}

	return ParseVersion(string(data))
}

func ParseVersion(version string) Version {
	versionStr := strings.TrimSpace(string(version))

	versionParts := strings.Split(versionStr, ".")

	if len(versionParts) != 2 {
		panic("Invalid version format")
	}

	major, err := strconv.Atoi(versionParts[0])
	if err != nil {
		panic(err)
	}

	minor, err := strconv.Atoi(versionParts[1])
	if err != nil {
		panic(err)
	}

	return Version{
		Major: major,
		Minor: minor,
	}
}
