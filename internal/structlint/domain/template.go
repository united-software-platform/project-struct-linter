package domain

// Template — курируемый эталон структуры проекта: версионированные данные, отдельные от логики
// проверки. Правила выражены полями данных, а не кодом (structure-template-model).
type Template struct {
	// SchemaVersion — версия схемы файла шаблона.
	SchemaVersion string
	// ID — идентификатор эталона (например, `go-standard`).
	ID string
	// Version — версия эталона в формате `MAJOR.MINOR.PATCH`.
	Version string
	// Language — язык/экосистема, для которой предназначен эталон.
	Language string
	// Detect — glob-маркеры авто-подбора: если все они присутствуют в структуре, эталон подходит.
	Detect []string
	// Required — правила обязательных путей.
	Required []PathRule
	// Forbidden — правила запрещённых путей.
	Forbidden []PathRule
	// Naming — правила именования файлов и каталогов в заданной области.
	Naming []NamingRule
	// Layers — правила разметки слоёв по путям.
	Layers []LayerRule
}

// TemplateRef — краткая ссылка на эталон для обнаружимости (list-templates).
type TemplateRef struct {
	// ID — идентификатор эталона.
	ID string
	// Version — версия эталона.
	Version string
	// Language — язык/экосистема эталона.
	Language string
}

// Ref возвращает краткую ссылку на эталон.
func (t Template) Ref() TemplateRef {
	return TemplateRef{ID: t.ID, Version: t.Version, Language: t.Language}
}

// PathRule — правило обязательного или запрещённого пути.
type PathRule struct {
	// Name — стабильный идентификатор правила для ссылки из нарушения.
	Name string
	// Kind — тип узла (file или dir).
	Kind Kind
	// Match — способ сопоставления (exact, glob, regex).
	Match Match
	// Pattern — образец пути.
	Pattern string
	// Min — минимальное число совпадений для glob-правила (по умолчанию 1).
	Min int
	// Reason — человекочитаемая причина, попадающая в нарушение.
	Reason string
}

// NamingRule — правило именования: имена в области Scope должны соответствовать Regex.
type NamingRule struct {
	// Name — стабильный идентификатор правила для ссылки из нарушения.
	Name string
	// Scope — glob-область, задающая проверяемые пути.
	Scope string
	// Kind — тип узла, к которому применяется правило; пусто — к файлам и каталогам.
	Kind Kind
	// Regex — регулярное выражение, которому должно соответствовать базовое имя узла.
	Regex string
	// Reason — человекочитаемая причина, попадающая в нарушение.
	Reason string
}

// LayerRule — правило разметки слоя по путям: узлы, отнесённые к слою селектором Selector,
// обязаны располагаться в пределах разрешённых шаблонов AllowedPaths. Анализ графа импортов
// в правило не входит.
type LayerRule struct {
	// Name — имя слоя и идентификатор правила для ссылки из нарушения.
	Name string
	// Selector — glob-селектор путей, относимых эталоном к слою.
	Selector string
	// AllowedPaths — glob-шаблоны разрешённого расположения путей слоя.
	AllowedPaths []string
	// Reason — человекочитаемая причина, попадающая в нарушение.
	Reason string
}
