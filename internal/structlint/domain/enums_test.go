package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"struct-linter/internal/structlint/domain"
)

// TestKindsЗамкнутоеМножество фиксирует допустимые значения Kind.
func TestKindsЗамкнутоеМножество(t *testing.T) {
	assert.ElementsMatch(t, []domain.Kind{domain.KindFile, domain.KindDir}, domain.Kinds())
}

// TestMatchesЗамкнутоеМножество фиксирует допустимые способы сопоставления.
func TestMatchesЗамкнутоеМножество(t *testing.T) {
	assert.ElementsMatch(t,
		[]domain.Match{domain.MatchExact, domain.MatchGlob, domain.MatchRegex},
		domain.Matches())
}

// TestCodesЗамкнутоеМножество фиксирует коды классов нарушений.
func TestCodesЗамкнутоеМножество(t *testing.T) {
	assert.ElementsMatch(t, []domain.Code{
		domain.CodeMissingRequired,
		domain.CodeForbiddenPresent,
		domain.CodeNamingViolation,
		domain.CodeLayerPathViolation,
	}, domain.Codes())
}

// TestStatusesЗамкнутоеМножество фиксирует статусы вердикта.
func TestStatusesЗамкнутоеМножество(t *testing.T) {
	assert.ElementsMatch(t,
		[]domain.Status{domain.StatusOK, domain.StatusViolations, domain.StatusError},
		domain.Statuses())
}

// TestЗначенияПеречисленийСтрочные проверяет строковые представления значений — они попадают
// в контракт (вердикт, шаблон) и не должны меняться незаметно.
func TestЗначенияПеречисленийСтрочные(t *testing.T) {
	assert.Equal(t, "file", string(domain.KindFile))
	assert.Equal(t, "dir", string(domain.KindDir))
	assert.Equal(t, "exact", string(domain.MatchExact))
	assert.Equal(t, "glob", string(domain.MatchGlob))
	assert.Equal(t, "regex", string(domain.MatchRegex))
	assert.Equal(t, "MISSING_REQUIRED", string(domain.CodeMissingRequired))
	assert.Equal(t, "FORBIDDEN_PRESENT", string(domain.CodeForbiddenPresent))
	assert.Equal(t, "NAMING_VIOLATION", string(domain.CodeNamingViolation))
	assert.Equal(t, "LAYER_PATH_VIOLATION", string(domain.CodeLayerPathViolation))
	assert.Equal(t, "ok", string(domain.StatusOK))
	assert.Equal(t, "violations", string(domain.StatusViolations))
	assert.Equal(t, "error", string(domain.StatusError))
}
