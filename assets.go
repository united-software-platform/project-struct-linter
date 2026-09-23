// Package assets встраивает в бинарь схемы контракта (schemas/) и курируемые эталоны
// (templates/), чтобы линтер был самодостаточным и запускался без файлов на диске (Standard 20):
// distroless-образ содержит только исполняемый файл, а данные читаются из встроенной ФС.
package assets

import "embed"

// Schemas — встроенные замкнутые схемы контракта (input, template, verdict).
//
//go:embed schemas/*.schema.json
var Schemas embed.FS

// Templates — встроенные курируемые эталоны структуры (templates/<id>/<version>/template.yaml).
//
//go:embed templates
var Templates embed.FS
