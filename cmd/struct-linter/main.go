// Команда struct-linter — CLI-линтер структуры проекта: проверяет структуру по курируемому
// эталону и печатает машиночитаемый вердикт с однозначным кодом возврата (0/1/2).
//
// Пакет main — единственный composition root (DI-005): здесь собирается граф зависимостей из
// встроенных схем и эталонов, а бизнес-логика живёт в internal/structlint.
package main

import (
	"fmt"
	"io/fs"
	"os"

	assets "struct-linter"
	"struct-linter/internal/structlint/application"
	"struct-linter/internal/structlint/infrastructure/clihandler"
	"struct-linter/internal/structlint/infrastructure/fsscanner"
	"struct-linter/internal/structlint/infrastructure/schemaguard"
	"struct-linter/internal/structlint/infrastructure/templaterepo"
)

// version — версия сборки линтера; переопределяется линковщиком (-ldflags).
var version = "dev"

func main() {
	os.Exit(run(os.Args[1:]))
}

// run собирает граф зависимостей и передаёт управление обработчику команд.
func run(args []string) int {
	schemasFS, err := fs.Sub(assets.Schemas, "schemas")
	if err != nil {
		fmt.Fprintf(os.Stderr, "struct-linter: встроенные схемы недоступны: %v\n", err)
		return clihandler.ExitError
	}
	templatesFS, err := fs.Sub(assets.Templates, "templates")
	if err != nil {
		fmt.Fprintf(os.Stderr, "struct-linter: встроенные эталоны недоступны: %v\n", err)
		return clihandler.ExitError
	}

	templateSchema, err := fs.ReadFile(schemasFS, "template.schema.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "struct-linter: схема эталона недоступна: %v\n", err)
		return clihandler.ExitError
	}
	guard, err := schemaguard.NewGuard(templateSchema)
	if err != nil {
		fmt.Fprintf(os.Stderr, "struct-linter: компиляция схемы эталона: %v\n", err)
		return clihandler.ExitError
	}

	repo := templaterepo.New(templatesFS, guard)
	scanner := fsscanner.New()
	validator := application.NewValidateStructure(repo, scanner)

	handler := clihandler.New(validator, repo, schemasFS, version, os.Stdout, os.Stderr)
	return handler.Run(args)
}
