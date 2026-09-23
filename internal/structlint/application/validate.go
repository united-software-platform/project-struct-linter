// Package application содержит правила приложения линтера: чистое ядро проверки структуры
// (Evaluate) над снимком дерева и эталоном и сценарий ValidateStructure, связывающий обход
// файловой системы и репозиторий эталонов с ядром. Слой зависит только от контрактов домена
// (CLAR-010) и не содержит деталей хранения и транспорта.
package application

import (
	"fmt"

	"struct-linter/internal/structlint/domain"
)

// AutoTemplate — значение выбора эталона, включающее авто-подбор по маркерам обнаружения.
const AutoTemplate = "auto"

// ValidateInput — вход сценария валидации: путь к проекту и выбор эталона.
type ValidateInput struct {
	// ProjectPath — путь к корню проверяемого проекта.
	ProjectPath string
	// TemplateID — идентификатор эталона либо AutoTemplate для авто-подбора.
	TemplateID string
	// TemplateVersion — версия эталона; пусто — последняя доступная.
	TemplateVersion string
}

// ValidateStructure — сценарий валидации структуры: обходит проект, выбирает эталон и применяет
// к снимку ядро проверки. Зависимости внедряются конструктором (DI-001, DI-004).
type ValidateStructure struct {
	templates domain.TemplateRepository
	scanner   domain.TreeScanner
}

// NewValidateStructure собирает сценарий валидации из контрактов домена.
func NewValidateStructure(templates domain.TemplateRepository, scanner domain.TreeScanner) *ValidateStructure {
	return &ValidateStructure{templates: templates, scanner: scanner}
}

// Execute обходит проект, выбирает эталон и возвращает вердикт проверки. Ошибки обхода и выбора
// эталона возвращаются доменными sentinel-ошибками для отображения в код возврата.
func (u *ValidateStructure) Execute(in ValidateInput) (domain.Verdict, error) {
	snapshot, err := u.scanner.Scan(in.ProjectPath)
	if err != nil {
		return domain.Verdict{}, err
	}
	tmpl, err := u.resolveTemplate(in, snapshot)
	if err != nil {
		return domain.Verdict{}, err
	}
	return Evaluate(snapshot, tmpl), nil
}

// resolveTemplate выбирает эталон по входу: авто-подбор по маркерам либо загрузка по id и версии.
func (u *ValidateStructure) resolveTemplate(in ValidateInput, snapshot domain.TreeSnapshot) (domain.Template, error) {
	if in.TemplateID == AutoTemplate || in.TemplateID == "" {
		return u.detect(snapshot)
	}
	return u.templates.Load(in.TemplateID, in.TemplateVersion)
}

// detect подбирает эталон, все маркеры обнаружения которого присутствуют в снимке. Репозиторий
// отдаёт эталоны по возрастанию версии, поэтому среди подходящих выбирается последний —
// наибольшая версия подходящего идентификатора.
func (u *ValidateStructure) detect(snapshot domain.TreeSnapshot) (domain.Template, error) {
	all, err := u.templates.List()
	if err != nil {
		return domain.Template{}, err
	}
	var chosen domain.Template
	found := false
	for _, tmpl := range all {
		if len(tmpl.Detect) > 0 && allMarkersPresent(tmpl.Detect, snapshot) {
			chosen = tmpl
			found = true
		}
	}
	if !found {
		return domain.Template{}, fmt.Errorf("%w: подходящий эталон не найден по маркерам обнаружения", domain.ErrTemplateNotFound)
	}
	return chosen, nil
}

// allMarkersPresent сообщает, присутствует ли в снимке хотя бы один узел на каждый маркер.
func allMarkersPresent(markers []string, snapshot domain.TreeSnapshot) bool {
	for _, marker := range markers {
		re, err := globToRegex(marker)
		if err != nil {
			return false
		}
		found := false
		for _, entry := range snapshot.Entries {
			if re.MatchString(entry.Path) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
