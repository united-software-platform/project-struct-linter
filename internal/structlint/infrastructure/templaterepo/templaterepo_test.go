package templaterepo_test

import (
	"fmt"
	"os"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"struct-linter/internal/structlint/application"
	"struct-linter/internal/structlint/domain"
	"struct-linter/internal/structlint/infrastructure/schemaguard"
	"struct-linter/internal/structlint/infrastructure/templaterepo"
)

const templateSchemaPath = "../../../../schemas/template.schema.json"

// tmplYAML собирает минимально валидный файл эталона с заданными id, версией и маркером.
func tmplYAML(id, version, detect string) string {
	return fmt.Sprintf(`schema_version: "1.0.0"
id: %s
version: "%s"
language: go
detect:
  - %s
required:
  - name: go-mod
    kind: file
    match: exact
    pattern: go.mod
    reason: нужен go.mod
`, id, version, detect)
}

func newRepo(t *testing.T, files fstest.MapFS) *templaterepo.Repository {
	t.Helper()
	schemaBytes, err := os.ReadFile(templateSchemaPath)
	require.NoError(t, err)
	guard, err := schemaguard.NewGuard(schemaBytes)
	require.NoError(t, err)
	return templaterepo.New(files, guard)
}

// fixtureFS собирает два id: sample с версиями 1.0.0 и 1.2.0, other с 1.0.0.
func fixtureFS() fstest.MapFS {
	return fstest.MapFS{
		"sample/1.0.0/template.yaml": {Data: []byte(tmplYAML("sample", "1.0.0", "go.mod"))},
		"sample/1.2.0/template.yaml": {Data: []byte(tmplYAML("sample", "1.2.0", "go.mod"))},
		"other/1.0.0/template.yaml":  {Data: []byte(tmplYAML("other", "1.0.0", "package.json"))},
	}
}

// TestLoadКонкретнаяВерсия проверяет выбор именно указанной версии.
func TestLoadКонкретнаяВерсия(t *testing.T) {
	repo := newRepo(t, fixtureFS())
	tmpl, err := repo.Load("sample", "1.0.0")
	require.NoError(t, err)
	assert.Equal(t, "1.0.0", tmpl.Version)
}

// TestLoadПоследняяВерсия проверяет выбор последней версии при пустой версии.
func TestLoadПоследняяВерсия(t *testing.T) {
	repo := newRepo(t, fixtureFS())
	tmpl, err := repo.Load("sample", "")
	require.NoError(t, err)
	assert.Equal(t, "1.2.0", tmpl.Version)
}

// TestLoadНесуществующийЭталон проверяет перевод отсутствия в domain.ErrTemplateNotFound.
func TestLoadНесуществующийЭталон(t *testing.T) {
	repo := newRepo(t, fixtureFS())
	_, err := repo.Load("missing", "9.9.9")
	assert.ErrorIs(t, err, domain.ErrTemplateNotFound)
}

// TestListВсеЭталоны проверяет перечень эталонов с идентификатором, версией и языком.
func TestListВсеЭталоны(t *testing.T) {
	repo := newRepo(t, fixtureFS())
	all, err := repo.List()
	require.NoError(t, err)
	require.Len(t, all, 3)

	refs := make([]domain.TemplateRef, 0, len(all))
	for _, tmpl := range all {
		refs = append(refs, tmpl.Ref())
	}
	assert.ElementsMatch(t, []domain.TemplateRef{
		{ID: "other", Version: "1.0.0", Language: "go"},
		{ID: "sample", Version: "1.0.0", Language: "go"},
		{ID: "sample", Version: "1.2.0", Language: "go"},
	}, refs)
}

// stubScanner — тестовый сканер, возвращающий заранее заданный снимок.
type stubScanner struct{ snap domain.TreeSnapshot }

func (s stubScanner) Scan(string) (domain.TreeSnapshot, error) { return s.snap, nil }

// TestАвтоПодборПоМаркерам проверяет выбор эталона по маркерам обнаружения через сценарий.
func TestАвтоПодборПоМаркерам(t *testing.T) {
	repo := newRepo(t, fixtureFS())
	scanner := stubScanner{snap: domain.TreeSnapshot{Entries: []domain.Entry{{Path: "go.mod"}}}}
	uc := application.NewValidateStructure(repo, scanner)

	v, err := uc.Execute(application.ValidateInput{ProjectPath: ".", TemplateID: application.AutoTemplate})
	require.NoError(t, err)
	// Маркер go.mod присутствует → выбран sample; среди версий — последняя (1.2.0).
	assert.Equal(t, "sample", v.Template.ID)
	assert.Equal(t, "1.2.0", v.Template.Version)
}

// TestАвтоПодборБезСовпадения проверяет ошибку вызова, когда маркеры не совпали.
func TestАвтоПодборБезСовпадения(t *testing.T) {
	repo := newRepo(t, fixtureFS())
	scanner := stubScanner{snap: domain.TreeSnapshot{Entries: []domain.Entry{{Path: "readme.md"}}}}
	uc := application.NewValidateStructure(repo, scanner)

	_, err := uc.Execute(application.ValidateInput{ProjectPath: ".", TemplateID: application.AutoTemplate})
	assert.ErrorIs(t, err, domain.ErrTemplateNotFound)
}
