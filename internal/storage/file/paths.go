package file

import "path/filepath"

func (s *Store) secretPath(name string) string {
	return filepath.Join(s.dir, name+".msk")
}
