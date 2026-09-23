package clihandler

import (
	"fmt"
	"io/fs"
	"strings"

	"struct-linter/internal/structlint/domain"
)

// templateRefDTO — краткая ссылка на эталон для JSON-вывода list-templates.
type templateRefDTO struct {
	ID       string `json:"id"`
	Version  string `json:"version"`
	Language string `json:"language"`
}

// templateListDTO — перечень эталонов для JSON-вывода.
type templateListDTO struct {
	Templates []templateRefDTO `json:"templates"`
}

// versionDTO — версия сборки для JSON-вывода.
type versionDTO struct {
	Version string `json:"version"`
}

// schemaFileFor отображает имя схемы контракта в имя встроенного файла.
func schemaFileFor(name string) (string, bool) {
	switch name {
	case "input", "template", "verdict":
		return name + ".schema.json", true
	default:
		return "", false
	}
}

// readSchema читает файл схемы из встроенной ФС схем.
func readSchema(schemas fs.FS, fileName string) ([]byte, error) {
	return fs.ReadFile(schemas, fileName)
}

// ensureTrailingNewline гарантирует завершающий перевод строки в выводе схемы.
func ensureTrailingNewline(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\n' {
		return data
	}
	return append(data, '\n')
}

// renderVerdictText отрисовывает вердикт в человекочитаемом виде.
func renderVerdictText(verdict domain.Verdict) string {
	var b strings.Builder
	fmt.Fprintf(&b, "status: %s\n", verdict.Status)
	if verdict.Template != (domain.TemplateRef{}) {
		fmt.Fprintf(&b, "template: %s %s (%s)\n", verdict.Template.ID, verdict.Template.Version, verdict.Template.Language)
	}
	if verdict.Error != nil {
		fmt.Fprintf(&b, "error: [%s] %s\n", verdict.Error.Code, verdict.Error.Message)
		return b.String()
	}
	if len(verdict.Violations) == 0 {
		b.WriteString("нарушений нет\n")
		return b.String()
	}
	for _, v := range verdict.Violations {
		fmt.Fprintf(&b, "- [%s] %s (%s): %s\n", v.Code, v.Path, v.Rule, v.Reason)
	}
	return b.String()
}
