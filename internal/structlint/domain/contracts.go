package domain

// TemplateRepository — контракт доступа к курируемым эталонам. Узкий контракт стороны-потребителя
// (STRUCT-001, STRUCT-002): реализация живёт в инфраструктуре, домен только объявляет.
type TemplateRepository interface {
	// Load возвращает эталон по идентификатору и версии; при пустой версии — последнюю доступную.
	// Отсутствие эталона возвращается как ErrTemplateNotFound.
	Load(id, version string) (Template, error)
	// List возвращает все доступные эталоны для обнаружимости и авто-подбора.
	List() ([]Template, error)
}

// TreeScanner — контракт построения снимка дерева проекта. Сканер читает только имена и пути,
// содержимое файлов не читает.
type TreeScanner interface {
	// Scan обходит дерево от корневого пути и возвращает снимок. Недоступный путь возвращается
	// как ErrProjectUnreadable.
	Scan(root string) (TreeSnapshot, error)
}
