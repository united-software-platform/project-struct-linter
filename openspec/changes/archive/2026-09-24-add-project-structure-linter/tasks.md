## 1. Каркас Go-модуля и домен

- [x] 1.1 Завести `go.mod` (модуль `struct-linter`); проверить `go build ./...` на пустом каркасе успешен
- [x] 1.2 Описать доменные типы `Template`, `Rule`, `Violation`, `Verdict`, `TreeSnapshot` и узкие контракты `TemplateRepository`, `TreeScanner` (1–3 метода, `STRUCT-001`) в `internal/structlint/domain` с package/doc comment (`GO-014`); проверить `go vet ./internal/structlint/domain`
- [x] 1.3 Зафиксировать замкнутые множества значений (`kind` ∈ {file,dir}, `match` ∈ {exact,glob,regex}, коды нарушений, статусы) как константы домена; проверить testify-тестом перечня допустимых значений (`GO-016`)
- [x] 1.4 Объявить доменные sentinel-ошибки `ErrTemplateNotFound`, `ErrProjectUnreadable`, `ErrTemplateInvalid` (`ERR-006`); проверить их инспекцию через `errors.Is`/`assert.ErrorIs` в тесте (`ERR-004`, `TEST-005`)

## 2. Контракт и схемы

- [x] 2.1 Написать `schemas/input.schema.json`, `schemas/template.schema.json`, `schemas/verdict.schema.json` с `additionalProperties: false`; проверить их самих на валидность JSON Schema
- [x] 2.2 Реализовать `schemaguard` (валидация файла шаблона по `template.schema.json` до применения); проверить тестами: неизвестное поле и недопустимый enum дают ошибку загрузки с указанием поля
- [x] 2.3 Встроить `schemas/` и `templates/` в бинарь через `embed`; проверить, что бинарь читает их без файлов на диске

## 3. Ядро валидации (application)

- [x] 3.1 Реализовать `ValidateStructure` (constructor injection интерфейсов, `DI-001`/`DI-004`) над `TreeSnapshot` + `Template`: правила required/forbidden; проверить testify-тестами коды `MISSING_REQUIRED` (включая недобор минимума glob) и `FORBIDDEN_PRESENT`
- [x] 3.2 Добавить правила naming и layers; проверить testify-тестами коды `NAMING_VIOLATION` и `LAYER_PATH_VIOLATION` с путём и причиной
- [x] 3.3 Обеспечить детерминизм: упорядочивание нарушений (правило, затем путь) и статус `ok`/`violations`; проверить тестом на повторный запуск (сравнение `JSONEq`, `TEST-008`) и на чистую фикстуру → `ok`

## 4. Инфраструктура и CLI

- [x] 4.1 Реализовать `fsscanner` (обход ФС от пути в `TreeSnapshot`, только имена/пути, без чтения содержимого); проверить тестом на фикстуре из `testdata/`
- [x] 4.2 Реализовать `templaterepo` (чтение `templates/<id>/<version>`, выбор последней версии, авто-подбор по `detect`); проверить тестами выбора версии и авто-подбора
- [x] 4.3 Реализовать `clihandler` с командами `validate`, `list-templates`, `schema`, `version` и JSON/text выводом; собрать граф зависимостей только в `cmd/struct-linter/main.go` (composition root, `DI-005`); проверить, что `validate` печатает вердикт по схеме
- [x] 4.4 Отобразить sentinel-ошибки в коды возврата `0`/`1`/`2` с полем `status`, логируя ошибку один раз на верхнем уровне (`ERR-005`/`ERR-010`); проверить тестами: ok→0, нарушения→1, `ErrTemplateNotFound`/`ErrProjectUnreadable`→2

## 5. Эталон и фикстуры

- [x] 5.1 Создать курируемый `templates/go-standard/1.0.0/template.yaml`; проверить его валидность по `template.schema.json`
- [x] 5.2 Создать фикстуры `testdata/`: чистый скелет (→ `ok`) и битые варианты на каждый класс нарушения; проверить `go test ./...` зелёным
- [x] 5.3 Добавить гейт калибровки: тест валидирует все `templates/**` по схеме и чистую фикстуру → `ok`; проверить прохождение теста

## 6. Поставка Docker и обёртка

- [x] 6.1 Написать `docker/struct-linter/Dockerfile` (сборка статического бинаря → distroless/scratch, непривилегированный пользователь по `DOCK-012`); проверить сборку образа
- [x] 6.2 Добавить сервис в `docker-compose.yml` и обёртку `tools/struct-linter/lint.sh` (`docker run --rm -v "$PWD":/work:ro ...`); проверить проброс кода возврата контейнера
- [x] 6.3 Добавить цели `Makefile` (сборка образа, запуск валидации) в норме `MAKE-001`; проверить `sh tools/make-lint/check.sh` зелёным

## 7. Гейт качества Go

- [x] 7.1 Прогнать `gofmt -l` (пусто) и `go vet ./...` — без замечаний (`GO-001`)
- [x] 7.2 Настроить и прогнать `golangci-lint run` (стиль, именование, модернизация — `GO-002`/`GO-004`/`GO-008`/`GO-009`/`GO-010`/`GO-017`); замечаний нет
- [x] 7.3 Прогнать `go test -race ./...` — все тесты зелёные, гонок нет (`SEC-010`)

## 8. Документация и приёмка

- [x] 8.1 Написать README линтера по `DOC-*` репо и структуре `GO-014` (`DOC-005`): контракт, примеры вызова слабой моделью, коды возврата, формат шаблона; проверить относительные ссылки (README без таблицы версий по `DOC-022`)
- [x] 8.2 Прогнать end-to-end по design.md: чистая фикстура→exit 0/`ok`, битая→exit 1 с путём и причиной, несуществующий шаблон→exit 2/`error`; `list-templates`/`schema` валидны по схемам
