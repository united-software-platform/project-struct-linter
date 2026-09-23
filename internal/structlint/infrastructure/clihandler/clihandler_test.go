package clihandler_test

import (
	"bytes"
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	assets "struct-linter"
	"struct-linter/internal/structlint/application"
	"struct-linter/internal/structlint/infrastructure/clihandler"
	"struct-linter/internal/structlint/infrastructure/fsscanner"
	"struct-linter/internal/structlint/infrastructure/schemaguard"
	"struct-linter/internal/structlint/infrastructure/templaterepo"
)

const (
	cleanFixture   = "../../../../testdata/clean"
	missingFixture = "../../../../testdata/broken-missing-required"
)

// harness — обработчик команд с перехваченными потоками вывода.
type harness struct {
	handler        *clihandler.Handler
	stdout, stderr *bytes.Buffer
	verdictSchema  *schemaguard.Compiled
}

// wire собирает обработчик над встроенными схемами и эталонами.
func wire(t *testing.T) *harness {
	t.Helper()
	schemasFS, err := fs.Sub(assets.Schemas, "schemas")
	require.NoError(t, err)
	templatesFS, err := fs.Sub(assets.Templates, "templates")
	require.NoError(t, err)

	templateSchema, err := fs.ReadFile(schemasFS, "template.schema.json")
	require.NoError(t, err)
	guard, err := schemaguard.NewGuard(templateSchema)
	require.NoError(t, err)

	verdictSchemaBytes, err := fs.ReadFile(schemasFS, "verdict.schema.json")
	require.NoError(t, err)
	verdictSchema, err := schemaguard.Compile("verdict.schema.json", verdictSchemaBytes)
	require.NoError(t, err)

	repo := templaterepo.New(templatesFS, guard)
	validator := application.NewValidateStructure(repo, fsscanner.New())

	out, errBuf := &bytes.Buffer{}, &bytes.Buffer{}
	handler := clihandler.New(validator, repo, schemasFS, "test", out, errBuf)
	return &harness{handler: handler, stdout: out, stderr: errBuf, verdictSchema: verdictSchema}
}

// TestValidateOkВозвращает0 проверяет: чистая фикстура → exit 0, вердикт валиден по схеме, status ok.
func TestValidateOkВозвращает0(t *testing.T) {
	h := wire(t)
	code := h.handler.Run([]string{"validate", "--template", "go-standard", "--project", cleanFixture})

	assert.Equal(t, clihandler.ExitOK, code)
	assert.NoError(t, h.verdictSchema.ValidateJSON(h.stdout.Bytes()))
	assert.Contains(t, h.stdout.String(), `"status": "ok"`)
}

// TestValidateНарушенияВозвращает1 проверяет: битая фикстура → exit 1, status violations.
func TestValidateНарушенияВозвращает1(t *testing.T) {
	h := wire(t)
	code := h.handler.Run([]string{"validate", "--template", "go-standard", "--project", missingFixture})

	assert.Equal(t, clihandler.ExitViolations, code)
	assert.NoError(t, h.verdictSchema.ValidateJSON(h.stdout.Bytes()))
	assert.Contains(t, h.stdout.String(), `"status": "violations"`)
}

// TestValidateБитыеФикстурыПоКлассам проверяет, что каждая битая фикстура даёт свой класс
// нарушения и код 1 (по одному классу на фикстуру).
func TestValidateБитыеФикстурыПоКлассам(t *testing.T) {
	cases := []struct {
		fixture string
		code    string
	}{
		{"../../../../testdata/broken-missing-required", "MISSING_REQUIRED"},
		{"../../../../testdata/broken-forbidden", "FORBIDDEN_PRESENT"},
		{"../../../../testdata/broken-naming", "NAMING_VIOLATION"},
		{"../../../../testdata/broken-layer", "LAYER_PATH_VIOLATION"},
	}
	for _, c := range cases {
		t.Run(c.code, func(t *testing.T) {
			h := wire(t)
			code := h.handler.Run([]string{"validate", "--template", "go-standard", "--project", c.fixture})

			assert.Equal(t, clihandler.ExitViolations, code)
			assert.NoError(t, h.verdictSchema.ValidateJSON(h.stdout.Bytes()))
			assert.Contains(t, h.stdout.String(), c.code)
		})
	}
}

// TestValidateЭталонНеНайденВозвращает2 проверяет: несуществующий эталон → exit 2, status error.
func TestValidateЭталонНеНайденВозвращает2(t *testing.T) {
	h := wire(t)
	code := h.handler.Run([]string{"validate", "--template", "no-such", "--project", cleanFixture})

	assert.Equal(t, clihandler.ExitError, code)
	assert.NoError(t, h.verdictSchema.ValidateJSON(h.stdout.Bytes()))
	assert.Contains(t, h.stdout.String(), `"status": "error"`)
	assert.Contains(t, h.stdout.String(), `"template_not_found"`)
}

// TestValidateПутьНедоступенВозвращает2 проверяет: недоступный путь → exit 2, code project_unreadable.
func TestValidateПутьНедоступенВозвращает2(t *testing.T) {
	h := wire(t)
	code := h.handler.Run([]string{"validate", "--template", "go-standard", "--project", "../../../../testdata/nope"})

	assert.Equal(t, clihandler.ExitError, code)
	assert.NoError(t, h.verdictSchema.ValidateJSON(h.stdout.Bytes()))
	assert.Contains(t, h.stdout.String(), `"project_unreadable"`)
}

// TestListTemplatesJSON проверяет непустой перечень эталонов и код 0.
func TestListTemplatesJSON(t *testing.T) {
	h := wire(t)
	code := h.handler.Run([]string{"list-templates"})

	assert.Equal(t, clihandler.ExitOK, code)
	assert.Contains(t, h.stdout.String(), `"id": "go-standard"`)
}

// TestSchemaVerdict проверяет выдачу схемы вердикта и код 0.
func TestSchemaVerdict(t *testing.T) {
	h := wire(t)
	code := h.handler.Run([]string{"schema", "--name", "verdict"})

	assert.Equal(t, clihandler.ExitOK, code)
	assert.Contains(t, h.stdout.String(), "verdict.schema.json")
}

// TestSchemaНеизвестнаяВозвращает2 проверяет ошибку на неизвестную схему.
func TestSchemaНеизвестнаяВозвращает2(t *testing.T) {
	h := wire(t)
	code := h.handler.Run([]string{"schema", "--name", "bogus"})
	assert.Equal(t, clihandler.ExitError, code)
}

// TestVersion проверяет вывод версии и код 0.
func TestVersion(t *testing.T) {
	h := wire(t)
	code := h.handler.Run([]string{"version"})

	assert.Equal(t, clihandler.ExitOK, code)
	assert.Contains(t, h.stdout.String(), "test")
}

// TestНеизвестнаяКомандаВозвращает2 проверяет ошибку на неизвестную команду.
func TestНеизвестнаяКомандаВозвращает2(t *testing.T) {
	h := wire(t)
	assert.Equal(t, clihandler.ExitError, h.handler.Run([]string{"frobnicate"}))
}
