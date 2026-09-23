package fsscanner_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"struct-linter/internal/structlint/domain"
	"struct-linter/internal/structlint/infrastructure/fsscanner"
)

// cleanFixture — путь к чистому скелету от каталога этого теста.
const cleanFixture = "../../../../testdata/clean"

// paths извлекает относительные пути записей снимка.
func paths(snap domain.TreeSnapshot) []string {
	out := make([]string, 0, len(snap.Entries))
	for _, e := range snap.Entries {
		out = append(out, e.Path)
	}
	return out
}

// TestScanСобираетОтносительныеПути проверяет обход фикстуры: относительные пути и признак каталога.
func TestScanСобираетОтносительныеПути(t *testing.T) {
	snap, err := fsscanner.New().Scan(cleanFixture)
	require.NoError(t, err)

	assert.ElementsMatch(t, []string{
		"go.mod",
		"cmd", "cmd/app", "cmd/app/main.go",
		"internal", "internal/sample", "internal/sample/domain",
		"internal/sample/domain/model.go",
	}, paths(snap))

	byPath := make(map[string]bool)
	for _, e := range snap.Entries {
		byPath[e.Path] = e.IsDir
	}
	assert.True(t, byPath["cmd"], "cmd должен быть каталогом")
	assert.False(t, byPath["go.mod"], "go.mod должен быть файлом")
}

// TestScanНедоступныйПуть проверяет перевод отсутствующего пути в domain.ErrProjectUnreadable.
func TestScanНедоступныйПуть(t *testing.T) {
	_, err := fsscanner.New().Scan("../../../../testdata/does-not-exist")
	assert.ErrorIs(t, err, domain.ErrProjectUnreadable)
}
