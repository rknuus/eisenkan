package harness

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// RecentStore is a simple file-backed list of recent board paths for tests.
type RecentStore struct {
	Path string
}

func NewRecentStore(dir string) *RecentStore {
	return &RecentStore{Path: filepath.Join(dir, "recent_boards.json")}
}

func (r *RecentStore) Add(path string) error {
	list, _ := r.List()
	// prepend unique
	out := []string{path}
	for _, p := range list {
		if p != path {
			out = append(out, p)
		}
	}
	// cap at 10
	if len(out) > 10 {
		out = out[:10]
	}
	return r.save(out)
}

func (r *RecentStore) List() ([]string, error) {
	b, err := os.ReadFile(r.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	var s []string
	if err := json.Unmarshal(b, &s); err != nil {
		return []string{}, nil
	}
	return s, nil
}

func (r *RecentStore) save(s []string) error {
	if err := os.MkdirAll(filepath.Dir(r.Path), 0o755); err != nil {
		return err
	}
	b, _ := json.MarshalIndent(s, "", "  ")
	return os.WriteFile(r.Path, b, 0o644)
}
