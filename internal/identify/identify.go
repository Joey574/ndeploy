package identify

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func FindDependencies(path string) ([]string, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// TODO: Add cyclic detection
	exp := regexp.MustCompile(`(?s)imports\s*=\s*\[(.*?)\];`)
	matches := exp.FindAllStringSubmatch(string(bytes), -1)
	if len(matches) == 0 {
		return nil, nil
	}

	var deps []string
	for i := range matches {
		cleaned := strings.TrimSpace(matches[i][1])
		deps = strings.Split(cleaned, "\n")

		// use len(deps) here to prevent scope from increasing while appending to deps
		for j := range len(deps) {
			deps[j] = strings.TrimSpace(deps[j])

			// hardware configuration is auto-generated and should not contain any new deps
			// additionally the import style used breaks this simple parsing strategy
			if deps[j] == "./hardware-configuration.nix" {
				continue
			}

			// dependencies can be absolute or relative paths
			var npath string
			if filepath.IsAbs(deps[j]) {
				npath = deps[j]
			} else {
				npath = filepath.Join(filepath.Dir(path), deps[j])
			}

			// recurse into dependencies
			ndeps, err := FindDependencies(npath)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", npath, err)
			}

			deps = append(deps, ndeps...)
		}
	}

	return deps, nil
}
