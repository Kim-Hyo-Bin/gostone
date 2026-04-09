package token

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/fernet/fernet-go"
)

// LoadFernetKeysFromRepo loads Keystone-style numbered key files (0, 1, 2, …) from dir.
// Keys are ordered by numeric id descending (highest = primary for encryption), matching
// keystone.common.fernet_utils.FernetUtils.load_keys.
func LoadFernetKeysFromRepo(dir string) ([]*fernet.Key, error) {
	ents, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("fernet key_repository %q: %w", dir, err)
	}
	type kv struct {
		id  int
		raw string
	}
	var keys []kv
	for _, e := range ents {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".tmp") {
			continue
		}
		id, err := strconv.Atoi(name)
		if err != nil {
			continue
		}
		path := filepath.Join(dir, name)
		b, err := os.ReadFile(path)
		if err != nil || len(b) == 0 {
			continue
		}
		keys = append(keys, kv{id: id, raw: strings.TrimSpace(string(b))})
	}
	if len(keys) == 0 {
		return nil, fmt.Errorf("fernet key_repository %q: no numeric key files", dir)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i].id > keys[j].id })
	out := make([]*fernet.Key, 0, len(keys))
	for _, k := range keys {
		fk, err := fernet.DecodeKey(strings.TrimSpace(k.raw))
		if err != nil {
			return nil, fmt.Errorf("fernet key file %d: %w", k.id, err)
		}
		out = append(out, fk)
	}
	return out, nil
}
