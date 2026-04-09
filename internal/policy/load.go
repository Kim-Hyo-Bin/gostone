package policy

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadFile reads an oslo.policy-style file and merges rules over Default().
// Supports .json (object of string rules), .yaml / .yml (flat string map), and
// unknown extensions by trying YAML then JSON.
func LoadFile(path string) (*Policy, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	overrides, err := parsePolicyBytes(raw, filepath.Ext(path))
	if err != nil {
		return nil, fmt.Errorf("parse policy file %s: %w", path, err)
	}
	p := Default()
	for k, v := range overrides {
		p.Rules[k] = v
	}
	return p, nil
}

func parsePolicyBytes(raw []byte, ext string) (map[string]string, error) {
	ext = strings.ToLower(ext)
	switch ext {
	case ".yaml", ".yml":
		return yamlToStringMap(raw)
	case ".json":
		return jsonToStringMap(raw)
	default:
		if o, err := yamlToStringMap(raw); err == nil && len(o) > 0 {
			return o, nil
		}
		return jsonToStringMap(raw)
	}
}

func jsonToStringMap(raw []byte) (map[string]string, error) {
	var overrides map[string]string
	if err := json.Unmarshal(raw, &overrides); err != nil {
		return nil, err
	}
	return overrides, nil
}

// yamlToStringMap loads Keystone/oslo.policy YAML: top-level keys are rule or action names.
func yamlToStringMap(raw []byte) (map[string]string, error) {
	var root map[string]interface{}
	if err := yaml.Unmarshal(raw, &root); err != nil {
		return nil, err
	}
	out := make(map[string]string, len(root))
	for k, v := range root {
		if k == "" {
			continue
		}
		out[k] = coercePolicyValue(v)
	}
	return out, nil
}

func coercePolicyValue(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case bool:
		if t {
			return "true"
		}
		return "false"
	case int:
		return fmt.Sprintf("%d", t)
	case int64:
		return fmt.Sprintf("%d", t)
	case float64:
		return fmt.Sprintf("%g", t)
	case nil:
		return ""
	default:
		return fmt.Sprint(t)
	}
}
