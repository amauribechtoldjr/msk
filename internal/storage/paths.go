package storage

import (
	"path/filepath"
	"strings"
)

func (s *Store) secretPath(name string) string {
	return filepath.Join(
		s.dir,
		strings.ToLower(name)+".msk",
	)
}
