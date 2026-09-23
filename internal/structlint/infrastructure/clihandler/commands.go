package clihandler

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"

	"struct-linter/internal/structlint/domain"
)

// runListTemplates печатает перечень доступных эталонов для обнаружимости.
func (h *Handler) runListTemplates(args []string) int {
	fsFlags := flag.NewFlagSet("list-templates", flag.ContinueOnError)
	fsFlags.SetOutput(h.stderr)
	format := fsFlags.String("format", "json", "формат вывода: json или text")
	if err := fsFlags.Parse(args); err != nil {
		return ExitError
	}

	all, err := h.templates.List()
	if err != nil {
		h.logf("%v", err)
		return ExitError
	}
	refs := make([]templateRefDTO, 0, len(all))
	for _, tmpl := range all {
		ref := tmpl.Ref()
		refs = append(refs, templateRefDTO{ID: ref.ID, Version: ref.Version, Language: ref.Language})
	}

	if *format == "text" {
		for _, ref := range refs {
			h.print(fmt.Sprintf("%s %s %s\n", ref.ID, ref.Version, ref.Language))
		}
		return ExitOK
	}
	data, err := json.MarshalIndent(templateListDTO{Templates: refs}, "", "  ")
	if err != nil {
		h.logf("сериализация перечня эталонов: %v", err)
		return ExitError
	}
	h.write(append(data, '\n'))
	return ExitOK
}

// runSchema выдаёт запрошенную схему контракта из встроенных схем.
func (h *Handler) runSchema(args []string) int {
	fsFlags := flag.NewFlagSet("schema", flag.ContinueOnError)
	fsFlags.SetOutput(h.stderr)
	name := fsFlags.String("name", "", "имя схемы: input, template или verdict")
	if err := fsFlags.Parse(args); err != nil {
		return ExitError
	}
	selected := *name
	if selected == "" && fsFlags.NArg() > 0 {
		selected = fsFlags.Arg(0)
	}
	fileName, ok := schemaFileFor(selected)
	if !ok {
		h.logf("неизвестная схема: %q (input, template или verdict)", selected)
		return ExitError
	}
	data, err := readSchema(h.schemas, fileName)
	if err != nil {
		h.logf("чтение схемы %s: %v", fileName, err)
		return ExitError
	}
	h.write(ensureTrailingNewline(data))
	return ExitOK
}

// runVersion печатает версию сборки линтера.
func (h *Handler) runVersion(args []string) int {
	fsFlags := flag.NewFlagSet("version", flag.ContinueOnError)
	fsFlags.SetOutput(h.stderr)
	format := fsFlags.String("format", "text", "формат вывода: json или text")
	if err := fsFlags.Parse(args); err != nil {
		return ExitError
	}
	if *format == "json" {
		data, err := json.MarshalIndent(versionDTO{Version: h.version}, "", "  ")
		if err != nil {
			h.logf("сериализация версии: %v", err)
			return ExitError
		}
		h.write(append(data, '\n'))
		return ExitOK
	}
	h.print(h.version + "\n")
	return ExitOK
}

// errorVerdict строит вердикт со статусом error по доменной ошибке.
func errorVerdict(err error) domain.Verdict {
	return domain.Verdict{
		SchemaVersion: domain.SchemaVersion,
		Status:        domain.StatusError,
		Violations:    nil,
		Error:         &domain.VerdictError{Code: errorCode(err), Message: err.Error()},
	}
}

// errorCode отображает доменные sentinel-ошибки в машинный код ошибки.
func errorCode(err error) string {
	switch {
	case errors.Is(err, domain.ErrTemplateNotFound):
		return "template_not_found"
	case errors.Is(err, domain.ErrProjectUnreadable):
		return "project_unreadable"
	case errors.Is(err, domain.ErrTemplateInvalid):
		return "template_invalid"
	default:
		return "error"
	}
}
