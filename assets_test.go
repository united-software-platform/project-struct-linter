package assets_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	assets "struct-linter"
	"struct-linter/internal/structlint/infrastructure/schemaguard"
)

// TestВстроенныеСхемыЧитаютсяИКомпилируются проверяет, что схемы доступны из встроенной ФС
// и являются валидными JSON Schema — без файлов на диске.
func TestВстроенныеСхемыЧитаютсяИКомпилируются(t *testing.T) {
	for _, name := range []string{"input.schema.json", "template.schema.json", "verdict.schema.json"} {
		data, err := assets.Schemas.ReadFile("schemas/" + name)
		require.NoError(t, err)
		_, err = schemaguard.Compile(name, data)
		assert.NoError(t, err)
	}
}

// TestВстроенныйЭталонПроходитСхему проверяет, что встроенный go-standard читается из ФС и
// проходит схему эталона.
func TestВстроенныйЭталонПроходитСхему(t *testing.T) {
	schemaBytes, err := assets.Schemas.ReadFile("schemas/template.schema.json")
	require.NoError(t, err)
	guard, err := schemaguard.NewGuard(schemaBytes)
	require.NoError(t, err)

	tmpl, err := assets.Templates.ReadFile("templates/go-standard/1.0.0/template.yaml")
	require.NoError(t, err)
	assert.NoError(t, guard.Validate(tmpl))
}
