package assets_test

import (
	"io/fs"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	assets "struct-linter"
	"struct-linter/internal/structlint/application"
	"struct-linter/internal/structlint/domain"
	"struct-linter/internal/structlint/infrastructure/fsscanner"
	"struct-linter/internal/structlint/infrastructure/schemaguard"
	"struct-linter/internal/structlint/infrastructure/templaterepo"
)

// TestКалибровкаВсеЭталоныПроходятСхему — гейт калибровки: каждый встроенный эталон валиден
// по схеме template.schema.json.
func TestКалибровкаВсеЭталоныПроходятСхему(t *testing.T) {
	schemaBytes, err := assets.Schemas.ReadFile("schemas/template.schema.json")
	require.NoError(t, err)
	guard, err := schemaguard.NewGuard(schemaBytes)
	require.NoError(t, err)

	count := 0
	walkErr := fs.WalkDir(assets.Templates, "templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, "template.yaml") {
			return nil
		}
		data, readErr := assets.Templates.ReadFile(path)
		require.NoError(t, readErr)
		assert.NoErrorf(t, guard.Validate(data), "эталон %s не проходит схему", path)
		count++
		return nil
	})
	require.NoError(t, walkErr)
	assert.Positive(t, count, "должен быть хотя бы один встроенный эталон")
}

// TestКалибровкаЧистаяФикстураOk — гейт калибровки: чистый скелет даёт вердикт ok по go-standard.
func TestКалибровкаЧистаяФикстураOk(t *testing.T) {
	schemaBytes, err := assets.Schemas.ReadFile("schemas/template.schema.json")
	require.NoError(t, err)
	guard, err := schemaguard.NewGuard(schemaBytes)
	require.NoError(t, err)

	templatesFS, err := fs.Sub(assets.Templates, "templates")
	require.NoError(t, err)
	repo := templaterepo.New(templatesFS, guard)
	uc := application.NewValidateStructure(repo, fsscanner.New())

	verdict, err := uc.Execute(application.ValidateInput{
		ProjectPath: "testdata/clean",
		TemplateID:  "go-standard",
	})
	require.NoError(t, err)
	assert.Equal(t, domain.StatusOK, verdict.Status)
	assert.Empty(t, verdict.Violations)
}
