package application

import (
	"fmt"
	"path"
	"regexp"
	"sort"

	"struct-linter/internal/structlint/domain"
)

// Evaluate — чистое ядро проверки: по снимку дерева и эталону возвращает детерминированный
// вердикт. Функция не обращается к файловой системе и транспорту (CLAR-010) и зависит только
// от входа, поэтому один и тот же вход всегда даёт один и тот же вердикт.
func Evaluate(snapshot domain.TreeSnapshot, tmpl domain.Template) domain.Verdict {
	var violations []domain.Violation
	violations = append(violations, checkRequired(snapshot, tmpl.Required)...)
	violations = append(violations, checkForbidden(snapshot, tmpl.Forbidden)...)
	violations = append(violations, checkNaming(snapshot, tmpl.Naming)...)
	violations = append(violations, checkLayers(snapshot, tmpl.Layers)...)

	sortViolations(violations)

	status := domain.StatusOK
	if len(violations) > 0 {
		status = domain.StatusViolations
	}
	return domain.Verdict{
		SchemaVersion: domain.SchemaVersion,
		Status:        status,
		Template:      tmpl.Ref(),
		Violations:    violations,
	}
}

// checkRequired проверяет наличие обязательных путей с учётом типа и способа сопоставления.
func checkRequired(snapshot domain.TreeSnapshot, rules []domain.PathRule) []domain.Violation {
	var out []domain.Violation
	for _, rule := range rules {
		count := countMatches(snapshot, rule)
		minMatches := rule.Min
		if minMatches <= 0 {
			minMatches = 1
		}
		if count >= minMatches {
			continue
		}
		reason := rule.Reason
		if rule.Match == domain.MatchGlob || rule.Match == domain.MatchRegex {
			reason = fmt.Sprintf("%s (найдено %d из %d)", rule.Reason, count, minMatches)
		}
		out = append(out, domain.Violation{
			Code:   domain.CodeMissingRequired,
			Path:   rule.Pattern,
			Rule:   rule.Name,
			Reason: reason,
		})
	}
	return out
}

// checkForbidden проверяет отсутствие запрещённых путей; на каждый найденный путь — нарушение.
func checkForbidden(snapshot domain.TreeSnapshot, rules []domain.PathRule) []domain.Violation {
	var out []domain.Violation
	for _, rule := range rules {
		matcher := ruleMatcher(rule)
		for _, entry := range snapshot.Entries {
			if !kindMatches(rule.Kind, entry) {
				continue
			}
			if matcher(entry.Path) {
				out = append(out, domain.Violation{
					Code:   domain.CodeForbiddenPresent,
					Path:   entry.Path,
					Rule:   rule.Name,
					Reason: rule.Reason,
				})
			}
		}
	}
	return out
}

// checkNaming проверяет соответствие базовых имён узлов регулярному выражению в области Scope.
func checkNaming(snapshot domain.TreeSnapshot, rules []domain.NamingRule) []domain.Violation {
	var out []domain.Violation
	for _, rule := range rules {
		scope, scopeErr := globToRegex(rule.Scope)
		name, nameErr := regexp.Compile(rule.Regex)
		if scopeErr != nil || nameErr != nil {
			continue
		}
		for _, entry := range snapshot.Entries {
			if !kindMatches(rule.Kind, entry) {
				continue
			}
			if !scope.MatchString(entry.Path) {
				continue
			}
			if name.MatchString(path.Base(entry.Path)) {
				continue
			}
			out = append(out, domain.Violation{
				Code:   domain.CodeNamingViolation,
				Path:   entry.Path,
				Rule:   rule.Name,
				Reason: rule.Reason,
			})
		}
	}
	return out
}

// checkLayers проверяет, что узлы, отнесённые к слою селектором, лежат в разрешённых путях.
func checkLayers(snapshot domain.TreeSnapshot, rules []domain.LayerRule) []domain.Violation {
	var out []domain.Violation
	for _, rule := range rules {
		selector, selErr := globToRegex(rule.Selector)
		if selErr != nil {
			continue
		}
		allowed := make([]*regexp.Regexp, 0, len(rule.AllowedPaths))
		for _, p := range rule.AllowedPaths {
			re, err := globToRegex(p)
			if err != nil {
				continue
			}
			allowed = append(allowed, re)
		}
		for _, entry := range snapshot.Entries {
			if !selector.MatchString(entry.Path) {
				continue
			}
			if matchesAny(allowed, entry.Path) {
				continue
			}
			out = append(out, domain.Violation{
				Code:   domain.CodeLayerPathViolation,
				Path:   entry.Path,
				Rule:   rule.Name,
				Reason: rule.Reason,
			})
		}
	}
	return out
}

// countMatches считает узлы, удовлетворяющие правилу пути с учётом типа.
func countMatches(snapshot domain.TreeSnapshot, rule domain.PathRule) int {
	matcher := ruleMatcher(rule)
	count := 0
	for _, entry := range snapshot.Entries {
		if !kindMatches(rule.Kind, entry) {
			continue
		}
		if matcher(entry.Path) {
			count++
		}
	}
	return count
}

// ruleMatcher возвращает предикат совпадения пути по способу сопоставления правила.
// Некомпилируемый образец даёт предикат, не совпадающий ни с чем: сам файл шаблона проходит
// схему до применения, поэтому ошибка компиляции здесь означает лишь безопасный отказ.
func ruleMatcher(rule domain.PathRule) func(string) bool {
	switch rule.Match {
	case domain.MatchExact:
		return func(p string) bool { return p == rule.Pattern }
	case domain.MatchGlob:
		re, err := globToRegex(rule.Pattern)
		if err != nil {
			return func(string) bool { return false }
		}
		return re.MatchString
	case domain.MatchRegex:
		re, err := regexp.Compile(rule.Pattern)
		if err != nil {
			return func(string) bool { return false }
		}
		return re.MatchString
	default:
		return func(string) bool { return false }
	}
}

// kindMatches проверяет соответствие типа узла ограничению правила; пустой Kind — без ограничения.
func kindMatches(kind domain.Kind, entry domain.Entry) bool {
	switch kind {
	case domain.KindFile:
		return !entry.IsDir
	case domain.KindDir:
		return entry.IsDir
	default:
		return true
	}
}

// matchesAny сообщает, совпадает ли путь хотя бы с одним из выражений.
func matchesAny(res []*regexp.Regexp, p string) bool {
	for _, re := range res {
		if re.MatchString(p) {
			return true
		}
	}
	return false
}

// sortViolations упорядочивает нарушения детерминированно: по правилу, затем по пути, затем по коду.
func sortViolations(violations []domain.Violation) {
	sort.SliceStable(violations, func(i, j int) bool {
		a, b := violations[i], violations[j]
		if a.Rule != b.Rule {
			return a.Rule < b.Rule
		}
		if a.Path != b.Path {
			return a.Path < b.Path
		}
		return a.Code < b.Code
	})
}
