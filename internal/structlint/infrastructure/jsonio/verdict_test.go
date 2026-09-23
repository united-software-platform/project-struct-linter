package jsonio_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"struct-linter/internal/structlint/application"
	"struct-linter/internal/structlint/domain"
	"struct-linter/internal/structlint/infrastructure/jsonio"
	"struct-linter/internal/structlint/infrastructure/schemaguard"
)

const verdictSchemaPath = "../../../../schemas/verdict.schema.json"

// sampleTemplate — эталон с правилом на каждый класс нарушения для проверок кодека.
func sampleTemplate() domain.Template {
	return domain.Template{
		ID: "go-standard", Version: "1.0.0", Language: "go",
		Required: []domain.PathRule{{
			Name: "cmd-dir", Kind: domain.KindDir, Match: domain.MatchExact,
			Pattern: "cmd", Reason: "нужен каталог cmd",
		}},
		Forbidden: []domain.PathRule{{
			Name: "no-tmp", Kind: domain.KindFile, Match: domain.MatchGlob,
			Pattern: "**/*.tmp", Reason: "нет .tmp",
		}},
	}
}

func brokenSnapshot() domain.TreeSnapshot {
	return domain.TreeSnapshot{Root: "/p", Entries: []domain.Entry{
		{Path: "go.mod", IsDir: false},
		{Path: "out.tmp", IsDir: false},
	}}
}

// TestВердиктСоответствуетСхеме сериализует вердикт и валидирует его по verdict.schema.json.
func TestВердиктСоответствуетСхеме(t *testing.T) {
	schemaBytes, err := os.ReadFile(filepath.Clean(verdictSchemaPath))
	require.NoError(t, err)
	compiled, err := schemaguard.Compile("verdict.schema.json", schemaBytes)
	require.NoError(t, err)

	v := application.Evaluate(brokenSnapshot(), sampleTemplate())
	data, err := jsonio.EncodeVerdict(v)
	require.NoError(t, err)

	assert.NoError(t, compiled.ValidateJSON(data))
}

// TestOkВердиктСоответствуетСхеме проверяет, что чистый вердикт (пустой список) валиден по схеме.
func TestOkВердиктСоответствуетСхеме(t *testing.T) {
	schemaBytes, err := os.ReadFile(filepath.Clean(verdictSchemaPath))
	require.NoError(t, err)
	compiled, err := schemaguard.Compile("verdict.schema.json", schemaBytes)
	require.NoError(t, err)

	clean := domain.TreeSnapshot{Root: "/p", Entries: []domain.Entry{
		{Path: "go.mod", IsDir: false},
		{Path: "cmd", IsDir: true},
	}}
	v := application.Evaluate(clean, sampleTemplate())
	require.Equal(t, domain.StatusOK, v.Status)

	data, err := jsonio.EncodeVerdict(v)
	require.NoError(t, err)
	assert.NoError(t, compiled.ValidateJSON(data))
}

// TestДетерминизмПовторногоЗапуска проверяет побайтную повторяемость вердикта (TEST-008).
func TestДетерминизмПовторногоЗапуска(t *testing.T) {
	first, err := jsonio.EncodeVerdict(application.Evaluate(brokenSnapshot(), sampleTemplate()))
	require.NoError(t, err)
	second, err := jsonio.EncodeVerdict(application.Evaluate(brokenSnapshot(), sampleTemplate()))
	require.NoError(t, err)

	assert.JSONEq(t, string(first), string(second))
	assert.Equal(t, string(first), string(second))
}
