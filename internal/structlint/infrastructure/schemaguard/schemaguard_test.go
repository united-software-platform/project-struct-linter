package schemaguard_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"struct-linter/internal/structlint/domain"
	"struct-linter/internal/structlint/infrastructure/schemaguard"
)

// schemasDir — путь к каталогу схем от каталога этого теста.
const schemasDir = "../../../../schemas"

func readSchema(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(schemasDir, name))
	require.NoError(t, err)
	return data
}

// TestСхемыКомпилируются проверяет, что все три файла схемы — валидные JSON Schema (2.1).
func TestСхемыКомпилируются(t *testing.T) {
	for _, name := range []string{"input.schema.json", "template.schema.json", "verdict.schema.json"} {
		t.Run(name, func(t *testing.T) {
			_, err := schemaguard.Compile(name, readSchema(t, name))
			assert.NoError(t, err)
		})
	}
}

// validTemplate — минимально корректный шаблон для позитивных проверок стража.
const validTemplate = `
schema_version: "1.0.0"
id: sample
version: "1.0.0"
language: go
required:
  - name: cmd-dir
    kind: dir
    match: exact
    pattern: cmd
    reason: точка входа обязательна
`

func newGuard(t *testing.T) *schemaguard.Guard {
	t.Helper()
	g, err := schemaguard.NewGuard(readSchema(t, "template.schema.json"))
	require.NoError(t, err)
	return g
}

// TestGuardПропускаетКорректныйШаблон проверяет, что валидный шаблон проходит стража.
func TestGuardПропускаетКорректныйШаблон(t *testing.T) {
	assert.NoError(t, newGuard(t).Validate([]byte(validTemplate)))
}

// TestGuardОтклоняетНеизвестноеПоле проверяет, что поле вне схемы блокирует загрузку с указанием поля.
func TestGuardОтклоняетНеизвестноеПоле(t *testing.T) {
	const withUnknown = validTemplate + "unknown_field: 42\n"
	err := newGuard(t).Validate([]byte(withUnknown))
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrTemplateInvalid)
	assert.Contains(t, err.Error(), "unknown_field")
}

// TestGuardОтклоняетНедопустимыйEnum проверяет, что значение вне перечисления блокирует загрузку
// с указанием поля.
func TestGuardОтклоняетНедопустимыйEnum(t *testing.T) {
	const badEnum = `
schema_version: "1.0.0"
id: sample
version: "1.0.0"
language: go
required:
  - name: cmd-dir
    kind: folder
    match: exact
    pattern: cmd
    reason: недопустимый kind
`
	err := newGuard(t).Validate([]byte(badEnum))
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrTemplateInvalid)
	assert.Contains(t, err.Error(), "kind")
}
