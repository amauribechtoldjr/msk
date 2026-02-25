package storage

import (
	"path/filepath"
	"strings"
)

func (s *Store) getFilePath(name string) string {
	return filepath.Join(
		s.dir,
		strings.ToLower(name)+".msk",
	)
}
