# Changelog

Все значимые изменения проекта фиксируются в этом файле.

Формат — [Keep a Changelog](https://keepachangelog.com/ru/1.1.0/); версионирование — по
[Semantic Versioning](https://semver.org/lang/ru/).

## [Unreleased]

## [1.0.0] - 2026-09-24

### Added

- CLI-линтер структуры проекта: команды `validate`, `list-templates`, `schema`, `version`.
- Замкнутый контракт вход/эталон/вердикт (JSON Schema, `additionalProperties: false`) и коды
  возврата `0`/`1`/`2` с дублирующим полем `status`.
- Движок правил: обязательные и запрещённые пути, именование, разметка слоёв по путям; вердикт с
  классифицированными нарушениями (путь и причина).
- Курируемый эталон `go-standard` `1.0.0`; авто-подбор эталона по маркерам (`--template auto`).
- Поставка Docker-образом (встроенные схемы и эталоны), обёртка `tools/struct-linter/lint.sh`,
  цели `Makefile`.

[Unreleased]: https://github.com/united-software-platform/project-struct-linter/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/united-software-platform/project-struct-linter/releases/tag/v1.0.0
