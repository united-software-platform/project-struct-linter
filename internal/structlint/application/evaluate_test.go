package application_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"struct-linter/internal/structlint/application"
	"struct-linter/internal/structlint/domain"
)

// snap собирает снимок дерева из пар «путь, признак каталога».
func snap(entries ...domain.Entry) domain.TreeSnapshot {
	return domain.TreeSnapshot{Root: "/project", Entries: entries}
}

func file(p string) domain.Entry { return domain.Entry{Path: p, IsDir: false} }
func dir(p string) domain.Entry  { return domain.Entry{Path: p, IsDir: true} }

// codesOf возвращает коды нарушений вердикта в порядке их следования.
func codesOf(v domain.Verdict) []domain.Code {
	out := make([]domain.Code, 0, len(v.Violations))
	for _, viol := range v.Violations {
		out = append(out, viol.Code)
	}
	return out
}

// TestОбязательныйПутьОтсутствует проверяет код MISSING_REQUIRED для exact-правила.
func TestОбязательныйПутьОтсутствует(t *testing.T) {
	tmpl := domain.Template{
		ID: "t", Version: "1.0.0", Language: "go",
		Required: []domain.PathRule{{
			Name: "cmd-dir", Kind: domain.KindDir, Match: domain.MatchExact,
			Pattern: "cmd", Reason: "нужен каталог cmd",
		}},
	}
	v := application.Evaluate(snap(file("go.mod")), tmpl)

	require.Len(t, v.Violations, 1)
	assert.Equal(t, domain.CodeMissingRequired, v.Violations[0].Code)
	assert.Equal(t, "cmd", v.Violations[0].Path)
	assert.Equal(t, "cmd-dir", v.Violations[0].Rule)
	assert.Equal(t, domain.StatusViolations, v.Status)
}

// TestОбязательныйGlobНеНабралМинимума проверяет недобор минимума совпадений glob.
func TestОбязательныйGlobНеНабралМинимума(t *testing.T) {
	tmpl := domain.Template{
		ID: "t", Version: "1.0.0", Language: "go",
		Required: []domain.PathRule{{
			Name: "entrypoint", Kind: domain.KindFile, Match: domain.MatchGlob,
			Pattern: "cmd/*/main.go", Min: 2, Reason: "нужны две точки входа",
		}},
	}
	v := application.Evaluate(snap(dir("cmd"), dir("cmd/app"), file("cmd/app/main.go")), tmpl)

	require.Len(t, v.Violations, 1)
	assert.Equal(t, domain.CodeMissingRequired, v.Violations[0].Code)
	assert.Equal(t, "entrypoint", v.Violations[0].Rule)
	assert.Contains(t, v.Violations[0].Reason, "1 из 2")
}

// TestОбязательныйGlobНаборМинимума проверяет отсутствие нарушения при достигнутом минимуме.
func TestОбязательныйGlobНаборМинимума(t *testing.T) {
	tmpl := domain.Template{
		ID: "t", Version: "1.0.0", Language: "go",
		Required: []domain.PathRule{{
			Name: "entrypoint", Kind: domain.KindFile, Match: domain.MatchGlob,
			Pattern: "cmd/*/main.go", Reason: "нужна точка входа",
		}},
	}
	v := application.Evaluate(snap(dir("cmd"), dir("cmd/app"), file("cmd/app/main.go")), tmpl)
	assert.Empty(t, v.Violations)
}

// TestЗапрещённыйПутьПрисутствует проверяет код FORBIDDEN_PRESENT.
func TestЗапрещённыйПутьПрисутствует(t *testing.T) {
	tmpl := domain.Template{
		ID: "t", Version: "1.0.0", Language: "go",
		Forbidden: []domain.PathRule{{
			Name: "no-tmp", Kind: domain.KindFile, Match: domain.MatchGlob,
			Pattern: "**/*.tmp", Reason: "временные файлы запрещены",
		}},
	}
	v := application.Evaluate(snap(file("go.mod"), file("build/out.tmp")), tmpl)

	require.Len(t, v.Violations, 1)
	assert.Equal(t, domain.CodeForbiddenPresent, v.Violations[0].Code)
	assert.Equal(t, "build/out.tmp", v.Violations[0].Path)
	assert.Equal(t, "временные файлы запрещены", v.Violations[0].Reason)
}

// TestИмяНеСоответствуетШаблону проверяет код NAMING_VIOLATION с путём и причиной.
func TestИмяНеСоответствуетШаблону(t *testing.T) {
	tmpl := domain.Template{
		ID: "t", Version: "1.0.0", Language: "go",
		Naming: []domain.NamingRule{{
			Name: "go-file-snake", Scope: "**/*.go", Kind: domain.KindFile,
			Regex: `^[a-z0-9_]+\.go$`, Reason: "имена Go-файлов в snake_case",
		}},
	}
	v := application.Evaluate(snap(file("internal/BadName.go"), file("internal/good_name.go")), tmpl)

	require.Len(t, v.Violations, 1)
	assert.Equal(t, domain.CodeNamingViolation, v.Violations[0].Code)
	assert.Equal(t, "internal/BadName.go", v.Violations[0].Path)
	assert.Equal(t, "go-file-snake", v.Violations[0].Rule)
}

// TestПутьВнеРазметкиСлоя проверяет код LAYER_PATH_VIOLATION с путём и причиной.
func TestПутьВнеРазметкиСлоя(t *testing.T) {
	tmpl := domain.Template{
		ID: "t", Version: "1.0.0", Language: "go",
		Layers: []domain.LayerRule{{
			Name: "domain", Selector: "**/domain/**",
			AllowedPaths: []string{"internal/*/domain/**"},
			Reason:       "доменные пакеты — в internal/<module>/domain",
		}},
	}
	v := application.Evaluate(snap(
		file("internal/sample/domain/model.go"),
		file("pkg/domain/leak.go"),
	), tmpl)

	require.Len(t, v.Violations, 1)
	assert.Equal(t, domain.CodeLayerPathViolation, v.Violations[0].Code)
	assert.Equal(t, "pkg/domain/leak.go", v.Violations[0].Path)
	assert.Equal(t, "domain", v.Violations[0].Rule)
}

// TestЧистаяСтруктураДаётOk проверяет статус ok и пустой список при отсутствии нарушений.
func TestЧистаяСтруктураДаётOk(t *testing.T) {
	tmpl := domain.Template{
		ID: "go-standard", Version: "1.0.0", Language: "go",
		Required: []domain.PathRule{{
			Name: "go-mod", Kind: domain.KindFile, Match: domain.MatchExact,
			Pattern: "go.mod", Reason: "нужен go.mod",
		}},
	}
	v := application.Evaluate(snap(file("go.mod"), dir("cmd")), tmpl)

	assert.Equal(t, domain.StatusOK, v.Status)
	assert.Empty(t, v.Violations)
	assert.Equal(t, "go-standard", v.Template.ID)
}

// TestПорядокНарушенийДетерминирован проверяет упорядочивание по правилу, затем по пути.
func TestПорядокНарушенийДетерминирован(t *testing.T) {
	tmpl := domain.Template{
		ID: "t", Version: "1.0.0", Language: "go",
		Forbidden: []domain.PathRule{{
			Name: "no-tmp", Kind: domain.KindFile, Match: domain.MatchGlob,
			Pattern: "**/*.tmp", Reason: "нет .tmp",
		}},
		Required: []domain.PathRule{{
			Name: "go-mod", Kind: domain.KindFile, Match: domain.MatchExact,
			Pattern: "go.mod", Reason: "нужен go.mod",
		}},
	}
	// Входные записи в «неудобном» порядке — вердикт всё равно упорядочен.
	v := application.Evaluate(snap(file("z/b.tmp"), file("a/a.tmp")), tmpl)

	require.Len(t, v.Violations, 3)
	// go-mod < no-tmp по имени правила; внутри no-tmp — по пути a/a.tmp < z/b.tmp.
	assert.Equal(t, []domain.Code{
		domain.CodeMissingRequired,
		domain.CodeForbiddenPresent,
		domain.CodeForbiddenPresent,
	}, codesOf(v))
	assert.Equal(t, "go-mod", v.Violations[0].Rule)
	assert.Equal(t, "a/a.tmp", v.Violations[1].Path)
	assert.Equal(t, "z/b.tmp", v.Violations[2].Path)
}
