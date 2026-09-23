// Package clihandler разбирает команды линтера (validate, list-templates, schema, version),
// вызывает сценарий приложения и печатает результат в JSON или человекочитаемом виде. Обработчик
// зависит от сценария и контрактов домена, а не от доменных объектов напрямую (CLAR-016), и
// логирует ошибку один раз на верхнем уровне (ERR-005, ERR-010), отображая её в код возврата 0/1/2.
package clihandler

import (
	"flag"
	"fmt"
	"io"
	"io/fs"

	"struct-linter/internal/structlint/application"
	"struct-linter/internal/structlint/domain"
	"struct-linter/internal/structlint/infrastructure/jsonio"
)

// Коды возврата линтера, дублируемые полем status вердикта.
const (
	// ExitOK — нарушений нет.
	ExitOK = 0
	// ExitViolations — найдены нарушения.
	ExitViolations = 1
	// ExitError — ошибка вызова или окружения.
	ExitError = 2
)

// Handler собирает зависимости команд и потоки вывода.
type Handler struct {
	validator *application.ValidateStructure
	templates domain.TemplateRepository
	schemas   fs.FS
	version   string
	stdout    io.Writer
	stderr    io.Writer
}

// New собирает обработчик команд из сценария, репозитория, встроенных схем и версии сборки.
func New(
	validator *application.ValidateStructure,
	templates domain.TemplateRepository,
	schemas fs.FS,
	version string,
	stdout, stderr io.Writer,
) *Handler {
	return &Handler{
		validator: validator,
		templates: templates,
		schemas:   schemas,
		version:   version,
		stdout:    stdout,
		stderr:    stderr,
	}
}

// Run разбирает команду и возвращает код возврата. Первый аргумент — имя подкоманды.
func (h *Handler) Run(args []string) int {
	if len(args) == 0 {
		h.logf("укажите команду: validate, list-templates, schema, version")
		return ExitError
	}
	command, rest := args[0], args[1:]
	switch command {
	case "validate":
		return h.runValidate(rest)
	case "list-templates":
		return h.runListTemplates(rest)
	case "schema":
		return h.runSchema(rest)
	case "version":
		return h.runVersion(rest)
	default:
		h.logf("неизвестная команда: %s", command)
		return ExitError
	}
}

// runValidate выполняет команду валидации структуры.
func (h *Handler) runValidate(args []string) int {
	fsFlags := flag.NewFlagSet("validate", flag.ContinueOnError)
	fsFlags.SetOutput(h.stderr)
	project := fsFlags.String("project", ".", "путь к корню проверяемого проекта")
	template := fsFlags.String("template", application.AutoTemplate, "идентификатор эталона либо auto")
	version := fsFlags.String("template-version", "", "версия эталона; пусто — последняя")
	format := fsFlags.String("format", "json", "формат вывода: json или text")
	if err := fsFlags.Parse(args); err != nil {
		return ExitError
	}
	if *format != "json" && *format != "text" {
		h.logf("недопустимый формат: %s (json или text)", *format)
		return ExitError
	}

	verdict, err := h.validator.Execute(application.ValidateInput{
		ProjectPath:     *project,
		TemplateID:      *template,
		TemplateVersion: *version,
	})
	if err != nil {
		return h.emitError(err, *format)
	}
	h.emitVerdict(verdict, *format)
	return exitForStatus(verdict.Status)
}

// emitVerdict печатает вердикт в выбранном формате.
func (h *Handler) emitVerdict(verdict domain.Verdict, format string) {
	if format == "text" {
		h.print(renderVerdictText(verdict))
		return
	}
	data, err := jsonio.EncodeVerdict(verdict)
	if err != nil {
		h.logf("сериализация вердикта: %v", err)
		return
	}
	h.write(data)
}

// emitError печатает вердикт-ошибку, логирует причину один раз и возвращает код 2.
func (h *Handler) emitError(err error, format string) int {
	verdict := errorVerdict(err)
	h.emitVerdict(verdict, format)
	h.logf("%v", err)
	return ExitError
}

// exitForStatus отображает статус вердикта в код возврата.
func exitForStatus(status domain.Status) int {
	switch status {
	case domain.StatusOK:
		return ExitOK
	case domain.StatusViolations:
		return ExitViolations
	default:
		return ExitError
	}
}

func (h *Handler) print(s string) { _, _ = fmt.Fprint(h.stdout, s) }
func (h *Handler) write(b []byte) { _, _ = h.stdout.Write(b) }
func (h *Handler) logf(format string, a ...any) {
	_, _ = fmt.Fprintf(h.stderr, "struct-linter: "+format+"\n", a...)
}
