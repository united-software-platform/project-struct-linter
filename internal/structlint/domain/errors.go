package domain

import "errors"

// Доменные sentinel-ошибки. Инфраструктура переводит свои ошибки в эти на границе слоя (ERR-006),
// а верхний уровень инспектирует их через errors.Is (ERR-004) и отображает в код возврата 2.
var (
	// ErrTemplateNotFound — эталон с заданным идентификатором или версией не найден.
	ErrTemplateNotFound = errors.New("template not found")
	// ErrProjectUnreadable — путь к проекту недоступен для обхода.
	ErrProjectUnreadable = errors.New("project unreadable")
	// ErrTemplateInvalid — файл шаблона не проходит замкнутую схему эталона.
	ErrTemplateInvalid = errors.New("template invalid")
)
