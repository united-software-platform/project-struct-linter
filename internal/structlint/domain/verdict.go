package domain

// SchemaVersion — версия схемы вердикта, выдаваемого линтером.
const SchemaVersion = "1.0.0"

// Violation — классифицированное нарушение: код класса, путь, ссылка на правило и причина.
// Нарушения без указания места (без пути) не допускаются.
type Violation struct {
	// Code — код класса нарушения.
	Code Code
	// Path — относительный путь, к которому относится нарушение.
	Path string
	// Rule — идентификатор нарушенного правила эталона.
	Rule string
	// Reason — человекочитаемая причина нарушения.
	Reason string
}

// VerdictError — сведения об ошибке вызова или окружения для вердикта со статусом error.
type VerdictError struct {
	// Code — машинный код ошибки (например, `template_not_found`).
	Code string
	// Message — человекочитаемое сообщение об ошибке.
	Message string
}

// Verdict — результат проверки структуры: статус, ссылка на применённый эталон и список
// нарушений; для статуса error — сведения об ошибке вместо нарушений.
type Verdict struct {
	// SchemaVersion — версия схемы вердикта.
	SchemaVersion string
	// Status — итоговый статус (ok, violations, error).
	Status Status
	// Template — применённый эталон; пусто при ошибке вызова до выбора эталона.
	Template TemplateRef
	// Violations — список нарушений; непустой при статусе violations, пустой при ok.
	Violations []Violation
	// Error — сведения об ошибке; заполнено только при статусе error.
	Error *VerdictError
}
