package conf

import (
	"bytes"
	"fmt"
	"os"

	"gopkg.in/ini.v1"
)

// Load reads INI configuration from resolved paths and returns merged settings with defaults.
func Load(opts Options) (*Config, error) {
	paths, err := resolveConfigPaths(opts)
	if err != nil {
		return nil, err
	}

	out := defaultConfig()
	if len(paths) == 0 {
		applyEnvOverrides(out)
		return out, nil
	}

	var buf bytes.Buffer
	for _, p := range paths {
		raw, err := os.ReadFile(p)
		if err != nil {
			return nil, fmt.Errorf("read config %s: %w", p, err)
		}
		buf.Write(raw)
		buf.WriteByte('\n')
	}

	f, err := ini.LoadSources(ini.LoadOptions{
		Insensitive: true,
	}, buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if err := f.MapTo(out); err != nil {
		return nil, fmt.Errorf("map config: %w", err)
	}

	applyEnvOverrides(out)
	return out, nil
}
