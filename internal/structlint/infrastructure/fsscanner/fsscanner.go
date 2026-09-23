// Package fsscanner строит снимок дерева проекта обходом файловой системы. Сканер читает только
// имена и относительные пути узлов и не открывает содержимое файлов (секреты проекта не читаются).
// Каталог системы контроля версий `.git` пропускается: он не относится к структуре проекта и его
// обход был бы дорогим и шумным.
package fsscanner

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"struct-linter/internal/structlint/domain"
)

// Scanner реализует domain.TreeScanner поверх файловой системы ОС.
type Scanner struct{}

var _ domain.TreeScanner = (*Scanner)(nil)

// New создаёт сканер файловой системы.
func New() *Scanner { return &Scanner{} }

// Scan обходит дерево от корневого пути и возвращает снимок с относительными путями узлов.
// Недоступный или несуществующий корень возвращается как domain.ErrProjectUnreadable.
func (s *Scanner) Scan(root string) (domain.TreeSnapshot, error) {
	info, err := filepath.Abs(root)
	if err != nil {
		return domain.TreeSnapshot{}, fmt.Errorf("%w: %v", domain.ErrProjectUnreadable, err)
	}
	root = info

	var entries []domain.Entry
	walkErr := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == root {
			return nil
		}
		if d.IsDir() && d.Name() == ".git" {
			return filepath.SkipDir
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		entries = append(entries, domain.Entry{
			Path:  filepath.ToSlash(rel),
			IsDir: d.IsDir(),
		})
		return nil
	})
	if walkErr != nil {
		return domain.TreeSnapshot{}, fmt.Errorf("%w: обход %s: %v", domain.ErrProjectUnreadable, root, walkErr)
	}
	return domain.TreeSnapshot{Root: root, Entries: entries}, nil
}
