package token

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fernet/fernet-go"
)

// WriteNextKeystoneFernetKey generates a 256-bit Fernet key and writes it to dir/N, where N is one greater
// than the highest existing numeric filename (0 if the directory has no numeric key files).
// The file is created atomically via N.tmp then rename, matching a typical Keystone key_repository layout.
func WriteNextKeystoneFernetKey(dir string) (nextID int, err error) {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return 0, fmt.Errorf("fernet key repository path is empty")
	}
	nextID, err = nextKeystoneFernetKeyID(dir)
	if err != nil {
		return 0, err
	}
	raw, err := generateFernetKeyString()
	if err != nil {
		return 0, err
	}
	tmpPath := filepath.Join(dir, fmt.Sprintf("%d.tmp", nextID))
	finalPath := filepath.Join(dir, strconv.Itoa(nextID))
	if err := os.WriteFile(tmpPath, []byte(raw+"\n"), 0o600); err != nil {
		return 0, fmt.Errorf("write %s: %w", tmpPath, err)
	}
	if err := os.Rename(tmpPath, finalPath); err != nil {
		_ = os.Remove(tmpPath)
		return 0, fmt.Errorf("rename to %s: %w", finalPath, err)
	}
	return nextID, nil
}

func nextKeystoneFernetKeyID(dir string) (int, error) {
	ents, err := os.ReadDir(dir)
	if err != nil {
		return 0, fmt.Errorf("fernet key_repository %q: %w", dir, err)
	}
	max := -1
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
		if id > max {
			max = id
		}
	}
	return max + 1, nil
}

func generateFernetKeyString() (string, error) {
	var k fernet.Key
	if err := k.Generate(); err != nil {
		return "", err
	}
	return k.Encode(), nil
}
