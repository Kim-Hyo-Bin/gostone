package conf

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// Options are CLI flags and env used to locate configuration files.
type Options struct {
	ConfigFile string // --config-file / -c
	ConfigDir  string // --config-dir
}

func fileExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && !st.IsDir()
}

func dirExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && st.IsDir()
}

// resolveConfigPaths returns files to read in order (later files override earlier keys).
// Discovery matches common Keystone deployments: explicit path, GOSTONE_CONFIG, then
// /etc/gostone/gostone.conf or ./gostone.conf, plus drop-ins from --config-dir,
// GOSTONE_CONFIG_DIR, or /etc/gostone/gostone.conf.d/*.conf when present.
func resolveConfigPaths(opts Options) ([]string, error) {
	var files []string

	switch {
	case opts.ConfigFile != "":
		if !fileExists(opts.ConfigFile) {
			return nil, fmt.Errorf("config file not found: %s", opts.ConfigFile)
		}
		files = append(files, opts.ConfigFile)
	case os.Getenv("GOSTONE_CONFIG") != "":
		p := os.Getenv("GOSTONE_CONFIG")
		if !fileExists(p) {
			return nil, fmt.Errorf("GOSTONE_CONFIG file not found: %s", p)
		}
		files = append(files, p)
	default:
		for _, p := range []string{
			"/etc/gostone/gostone.conf",
			"gostone.conf",
		} {
			if fileExists(p) {
				files = append(files, p)
				break
			}
		}
	}

	frags, err := fragmentConfFiles(opts)
	if err != nil {
		return nil, err
	}
	files = append(files, frags...)
	return files, nil
}

func fragmentConfFiles(opts Options) ([]string, error) {
	dir := opts.ConfigDir
	if dir == "" {
		dir = os.Getenv("GOSTONE_CONFIG_DIR")
	}
	if dir == "" && dirExists("/etc/gostone/gostone.conf.d") {
		dir = "/etc/gostone/gostone.conf.d"
	}
	if dir == "" {
		return nil, nil
	}
	if !dirExists(dir) {
		return nil, fmt.Errorf("config directory not found: %s", dir)
	}
	matches, err := filepath.Glob(filepath.Join(dir, "*.conf"))
	if err != nil {
		return nil, err
	}
	sort.Strings(matches)
	return matches, nil
}
