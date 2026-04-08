package policy

import (
	"encoding/json"
	"fmt"
	"os"
)

// LoadFile reads a JSON object of rule strings and merges them over Default().
func LoadFile(path string) (*Policy, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var overrides map[string]string
	if err := json.Unmarshal(raw, &overrides); err != nil {
		return nil, fmt.Errorf("parse policy json: %w", err)
	}
	p := Default()
	for k, v := range overrides {
		p.Rules[k] = v
	}
	return p, nil
}
