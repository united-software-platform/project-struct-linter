package domain

// Kind — тип узла структуры, к которому относится правило пути.
type Kind string

const (
	// KindFile — правило относится к файлу.
	KindFile Kind = "file"
	// KindDir — правило относится к каталогу.
	KindDir Kind = "dir"
)

// Match — способ сопоставления пути с образцом правила.
type Match string

const (
	// MatchExact — точное совпадение относительного пути.
	MatchExact Match = "exact"
	// MatchGlob — совпадение по glob-шаблону (`*`, `**`, `?`).
	MatchGlob Match = "glob"
	// MatchRegex — совпадение по регулярному выражению.
	MatchRegex Match = "regex"
)

// Code — класс нарушения в вердикте.
type Code string

const (
	// CodeMissingRequired — обязательный путь отсутствует либо glob не набрал минимума.
	CodeMissingRequired Code = "MISSING_REQUIRED"
	// CodeForbiddenPresent — в структуре присутствует запрещённый эталоном путь.
	CodeForbiddenPresent Code = "FORBIDDEN_PRESENT"
	// CodeNamingViolation — имя в области правила не соответствует его регулярному выражению.
	CodeNamingViolation Code = "NAMING_VIOLATION"
	// CodeLayerPathViolation — путь нарушает объявленную для слоя разметку.
	CodeLayerPathViolation Code = "LAYER_PATH_VIOLATION"
)

// Status — итоговый статус вердикта.
type Status string

const (
	// StatusOK — нарушений не найдено.
	StatusOK Status = "ok"
	// StatusViolations — найдено хотя бы одно нарушение.
	StatusViolations Status = "violations"
	// StatusError — ошибка вызова или окружения (шаблон не найден, путь недоступен, шаблон невалиден).
	StatusError Status = "error"
)

// Kinds возвращает замкнутое множество допустимых значений Kind.
func Kinds() []Kind { return []Kind{KindFile, KindDir} }

// Matches возвращает замкнутое множество допустимых значений Match.
func Matches() []Match { return []Match{MatchExact, MatchGlob, MatchRegex} }

// Codes возвращает замкнутое множество кодов нарушений.
func Codes() []Code {
	return []Code{CodeMissingRequired, CodeForbiddenPresent, CodeNamingViolation, CodeLayerPathViolation}
}

// Statuses возвращает замкнутое множество статусов вердикта.
func Statuses() []Status { return []Status{StatusOK, StatusViolations, StatusError} }
