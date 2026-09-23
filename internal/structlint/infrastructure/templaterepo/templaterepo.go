// Package templaterepo читает курируемые эталоны из файловой системы (встроенной или дисковой):
// каталоги templates/<id>/<version>/template.yaml. Каждый файл проверяется схемой эталона до
// разбора (schemaguard), а инфраструктурные ошибки переводятся в доменные sentinel (ERR-006).
package templaterepo

import (
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	"struct-linter/internal/structlint/domain"
	"struct-linter/internal/structlint/infrastructure/schemaguard"
)

// templateFileName — имя файла эталона в каталоге версии.
const templateFileName = "template.yaml"

// Repository читает эталоны из fs.FS, корень которого содержит каталоги идентификаторов.
type Repository struct {
	fsys  fs.FS
	guard *schemaguard.Guard
}

var _ domain.TemplateRepository = (*Repository)(nil)

// New собирает репозиторий эталонов над файловой системой и стражем схемы.
func New(fsys fs.FS, guard *schemaguard.Guard) *Repository {
	return &Repository{fsys: fsys, guard: guard}
}

// Load возвращает эталон по идентификатору и версии; при пустой версии — последнюю доступную.
func (r *Repository) Load(id, version string) (domain.Template, error) {
	if version == "" {
		latest, err := r.latestVersion(id)
		if err != nil {
			return domain.Template{}, err
		}
		version = latest
	}
	file := path.Join(id, version, templateFileName)
	data, err := fs.ReadFile(r.fsys, file)
	if err != nil {
		return domain.Template{}, fmt.Errorf("%w: эталон %s версии %s", domain.ErrTemplateNotFound, id, version)
	}
	return r.parse(data)
}

// List возвращает все доступные эталоны для обнаружимости и авто-подбора.
func (r *Repository) List() ([]domain.Template, error) {
	ids, err := fs.ReadDir(r.fsys, ".")
	if err != nil {
		return nil, fmt.Errorf("%w: перечень эталонов: %v", domain.ErrTemplateNotFound, err)
	}
	var out []domain.Template
	for _, idEntry := range ids {
		if !idEntry.IsDir() {
			continue
		}
		versions, verErr := r.versions(idEntry.Name())
		if verErr != nil {
			return nil, verErr
		}
		for _, version := range versions {
			tmpl, loadErr := r.Load(idEntry.Name(), version)
			if loadErr != nil {
				return nil, loadErr
			}
			out = append(out, tmpl)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].ID != out[j].ID {
			return out[i].ID < out[j].ID
		}
		return compareVersions(out[i].Version, out[j].Version) < 0
	})
	return out, nil
}

// versions перечисляет версии эталона в порядке возрастания.
func (r *Repository) versions(id string) ([]string, error) {
	entries, err := fs.ReadDir(r.fsys, id)
	if err != nil {
		return nil, fmt.Errorf("%w: версии эталона %s", domain.ErrTemplateNotFound, id)
	}
	var versions []string
	for _, e := range entries {
		if e.IsDir() {
			versions = append(versions, e.Name())
		}
	}
	sort.Slice(versions, func(i, j int) bool { return compareVersions(versions[i], versions[j]) < 0 })
	return versions, nil
}

// latestVersion возвращает наибольшую версию эталона.
func (r *Repository) latestVersion(id string) (string, error) {
	versions, err := r.versions(id)
	if err != nil {
		return "", err
	}
	if len(versions) == 0 {
		return "", fmt.Errorf("%w: у эталона %s нет версий", domain.ErrTemplateNotFound, id)
	}
	return versions[len(versions)-1], nil
}

// parse валидирует файл схемой и разбирает его в доменный эталон.
func (r *Repository) parse(data []byte) (domain.Template, error) {
	if err := r.guard.Validate(data); err != nil {
		return domain.Template{}, err
	}
	var file templateFile
	if err := yaml.Unmarshal(data, &file); err != nil {
		return domain.Template{}, fmt.Errorf("%w: разбор шаблона: %v", domain.ErrTemplateInvalid, err)
	}
	return file.toDomain(), nil
}

// compareVersions сравнивает версии `X.Y.Z` покомпонентно: -1, 0 или 1.
func compareVersions(a, b string) int {
	as, bs := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < len(as) || i < len(bs); i++ {
		av, bv := component(as, i), component(bs, i)
		if av != bv {
			if av < bv {
				return -1
			}
			return 1
		}
	}
	return 0
}

// component возвращает числовое значение i-го компонента версии или 0, если его нет.
func component(parts []string, i int) int {
	if i >= len(parts) {
		return 0
	}
	n, err := strconv.Atoi(parts[i])
	if err != nil {
		return 0
	}
	return n
}
