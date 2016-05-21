package module

import (
	"os"
	"path/filepath"
	"strings"
)

// Extension is the extension used by module files
const Extension = ".hcl"

// DiscoveredRegistry type contains discovered modules as returned by the
// Discover() function.
// Keys of the map are the module names and their values are the
// absolute path to the discovered module files.
type DiscoveredRegistry map[string]string

// LoadedRegistry type is a map which keys are the
// discovered modules from a given module path and their
// values are the actual loaded modules.
type LoadedRegistry map[string]*Module

// Discover is used to discover valid modules in a given module path
func Discover(root string) (DiscoveredRegistry, error) {
	registry := make(DiscoveredRegistry)

	// Module walker function
	walker := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directory entries
		if info.IsDir() {
			return nil
		}

		// Skip files which don't appear to be valid module files
		fileExt := filepath.Ext(info.Name())
		if fileExt != Extension {
			return nil
		}

		// Remove the root path portion from the discovered module file,
		// remove the module file extension and register the module
		moduleFileWithExt := strings.TrimPrefix(path, root)
		moduleNameWithExt := strings.TrimSuffix(moduleFileWithExt, fileExt)
		moduleName := strings.Trim(moduleNameWithExt, string(os.PathSeparator))
		absPath, err := filepath.Abs(path)
		if err != nil {
			return err
		}

		registry[moduleName] = absPath

		return nil
	}

	return registry, filepath.Walk(root, walker)
}

// DiscoverAndLoad discovers valid modules from a given
// module path and attemps to load each valid module file.
func DiscoverAndLoad(config *Config) (LoadedRegistry, error) {
	registry := make(LoadedRegistry)

	discovered, err := Discover(config.Path)
	if err != nil {
		return registry, err
	}

	for n, p := range discovered {
		f, err := os.Open(p)
		if err != nil {
			return registry, err
		}

		m, err := Load(n, config, f)
		if err != nil {
			return registry, err
		}
		registry[n] = m
		f.Close()
	}

	return registry, nil
}
