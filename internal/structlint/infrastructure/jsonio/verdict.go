// Package jsonio сериализует вердикт линтера в JSON по схеме verdict.schema.json: имена полей
// в snake_case, список нарушений присутствует всегда (пустой при статусе ok). Пакет — граница
// транспорта: он переводит доменный тип в формат контракта, не внося логики проверки.
package jsonio

import (
	"encoding/json"

	"struct-linter/internal/structlint/domain"
)

type templateDTO struct {
	ID       string `json:"id"`
	Version  string `json:"version"`
	Language string `json:"language"`
}

type violationDTO struct {
	Code   string `json:"code"`
	Path   string `json:"path"`
	Rule   string `json:"rule"`
	Reason string `json:"reason"`
}

type errorDTO struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type verdictDTO struct {
	SchemaVersion string         `json:"schema_version"`
	Status        string         `json:"status"`
	Template      *templateDTO   `json:"template,omitempty"`
	Violations    []violationDTO `json:"violations"`
	Error         *errorDTO      `json:"error,omitempty"`
}

// toDTO переводит доменный вердикт в структуру контракта.
func toDTO(v domain.Verdict) verdictDTO {
	dto := verdictDTO{
		SchemaVersion: v.SchemaVersion,
		Status:        string(v.Status),
		Violations:    make([]violationDTO, 0, len(v.Violations)),
	}
	if v.SchemaVersion == "" {
		dto.SchemaVersion = domain.SchemaVersion
	}
	if v.Template != (domain.TemplateRef{}) {
		dto.Template = &templateDTO{
			ID:       v.Template.ID,
			Version:  v.Template.Version,
			Language: v.Template.Language,
		}
	}
	for _, viol := range v.Violations {
		dto.Violations = append(dto.Violations, violationDTO{
			Code:   string(viol.Code),
			Path:   viol.Path,
			Rule:   viol.Rule,
			Reason: viol.Reason,
		})
	}
	if v.Error != nil {
		dto.Error = &errorDTO{Code: v.Error.Code, Message: v.Error.Message}
	}
	return dto
}

// EncodeVerdict сериализует вердикт в отступованный JSON по схеме вердикта.
func EncodeVerdict(v domain.Verdict) ([]byte, error) {
	data, err := json.MarshalIndent(toDTO(v), "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}
