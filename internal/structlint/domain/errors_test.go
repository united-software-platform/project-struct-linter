package domain_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"struct-linter/internal/structlint/domain"
)

// TestSentinelОшибкиИнспектируютсяIs проверяет, что обёрнутые доменные ошибки инспектируются
// через errors.Is (ERR-004, TEST-005).
func TestSentinelОшибкиИнспектируютсяIs(t *testing.T) {
	cases := []struct {
		name string
		err  error
	}{
		{"template not found", domain.ErrTemplateNotFound},
		{"project unreadable", domain.ErrProjectUnreadable},
		{"template invalid", domain.ErrTemplateInvalid},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			wrapped := fmt.Errorf("граница слоя: %w", c.err)
			assert.ErrorIs(t, wrapped, c.err)
		})
	}
}

// TestSentinelОшибкиРазличимы проверяет, что доменные ошибки не путаются одна с другой.
func TestSentinelОшибкиРазличимы(t *testing.T) {
	assert.NotErrorIs(t, domain.ErrTemplateNotFound, domain.ErrProjectUnreadable)
	assert.NotErrorIs(t, domain.ErrProjectUnreadable, domain.ErrTemplateInvalid)
	assert.NotErrorIs(t, domain.ErrTemplateInvalid, domain.ErrTemplateNotFound)
}
