// Package schemaguard валидирует данные по замкнутым JSON Schema контракта до их применения:
// файл шаблона проверяется схемой эталона, и поле или значение вне схемы блокирует загрузку,
// а не игнорируется молча (structure-template-model). Пакет — граница слоя: ошибку схемы он
// переводит в доменную ErrTemplateInvalid (ERR-006).
package schemaguard

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"gopkg.in/yaml.v3"

	"struct-linter/internal/structlint/domain"
)

// Compiled — скомпилированная замкнутая схема, готовая проверять экземпляры.
type Compiled struct {
	schema *jsonschema.Schema
}

// Compile компилирует JSON Schema из её байтов. Имя используется как идентификатор ресурса
// в сообщениях об ошибках.
func Compile(name string, schemaJSON []byte) (*Compiled, error) {
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(schemaJSON))
	if err != nil {
		return nil, fmt.Errorf("разбор схемы %s: %w", name, err)
	}
	c := jsonschema.NewCompiler()
	if err := c.AddResource(name, doc); err != nil {
		return nil, fmt.Errorf("регистрация схемы %s: %w", name, err)
	}
	schema, err := c.Compile(name)
	if err != nil {
		return nil, fmt.Errorf("компиляция схемы %s: %w", name, err)
	}
	return &Compiled{schema: schema}, nil
}

// ValidateJSON проверяет JSON-документ по схеме. Ошибка валидации возвращается как есть,
// с указанием места несоответствия.
func (c *Compiled) ValidateJSON(data []byte) error {
	inst, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("разбор экземпляра: %w", err)
	}
	if err := c.schema.Validate(inst); err != nil {
		return err
	}
	return nil
}

// Guard проверяет файл шаблона по замкнутой схеме эталона перед его применением.
type Guard struct {
	compiled *Compiled
}

// NewGuard компилирует схему эталона и возвращает страж загрузки шаблонов.
func NewGuard(templateSchema []byte) (*Guard, error) {
	compiled, err := Compile("template.schema.json", templateSchema)
	if err != nil {
		return nil, err
	}
	return &Guard{compiled: compiled}, nil
}

// Validate проверяет YAML-файл шаблона по схеме эталона. Неизвестное поле или недопустимое
// значение перечисления возвращаются как domain.ErrTemplateInvalid с указанием поля и причины.
func (g *Guard) Validate(templateYAML []byte) error {
	var value any
	if err := yaml.Unmarshal(templateYAML, &value); err != nil {
		return fmt.Errorf("%w: разбор yaml: %v", domain.ErrTemplateInvalid, err)
	}
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("%w: приведение к json: %v", domain.ErrTemplateInvalid, err)
	}
	if err := g.compiled.ValidateJSON(jsonBytes); err != nil {
		// Сообщение santhosh включает instance location (`at '/required/0/kind': ...`) —
		// это и есть указание поля, требуемое контрактом загрузки шаблона.
		return fmt.Errorf("%w: %w", domain.ErrTemplateInvalid, err)
	}
	return nil
}
