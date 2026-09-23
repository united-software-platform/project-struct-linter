# Go Rules — стандартизированный свод правил

Набор правил разработки на Go, оформленный по единому шаблону
(`process/command_scenario/template-rules.md`). Каждый документ содержит
таблицу правил с уникальными ID, подробное описание, примеры и версионирование.

## Как ссылаться на правило

Формат ID внутри документа — `PREFIX-NNN` (например, `ERR-002`, `CONC-007`).
Имя файла — `GO-NNN-<ТЕМА>.md`. Ссылка на конкретное правило: `GO-005-ERROR-HANDLING → ERR-002`.

## Каталог правил

| Файл | Тема | Префикс ID |
|------|------|------------|
| [GO-001-CODE-STYLE](GO-001-CODE-STYLE.md) | Стиль кода и форматирование | `STYLE` |
| [GO-002-NAMING](GO-002-NAMING.md) | Соглашения об именовании | `NAME` |
| [GO-003-STRUCTS-INTERFACES](GO-003-STRUCTS-INTERFACES.md) | Структуры и интерфейсы | `STRUCT` |
| [GO-004-DATA-STRUCTURES](GO-004-DATA-STRUCTURES.md) | Структуры данных | `DATA` |
| [GO-005-ERROR-HANDLING](GO-005-ERROR-HANDLING.md) | Обработка ошибок | `ERR` |
| [GO-006-CONTEXT](GO-006-CONTEXT.md) | Работа с context.Context | `CTX` |
| [GO-007-CONCURRENCY](GO-007-CONCURRENCY.md) | Конкурентность | `CONC` |
| [GO-008-SAFETY](GO-008-SAFETY.md) | Защитное программирование | `SAFE` |
| [GO-009-PERFORMANCE](GO-009-PERFORMANCE.md) | Производительность | `PERF` |
| [GO-010-DESIGN-PATTERNS](GO-010-DESIGN-PATTERNS.md) | Идиомы и паттерны проектирования | `PAT` |
| [GO-011-DEPENDENCY-INJECTION](GO-011-DEPENDENCY-INJECTION.md) | Внедрение зависимостей | `DI` |
| [GO-012-DATABASE](GO-012-DATABASE.md) | Работа с базами данных | `DB` |
| [GO-013-GRPC](GO-013-GRPC.md) | gRPC | `GRPC` |
| [GO-014-DOCUMENTATION](GO-014-DOCUMENTATION.md) | Документирование | `DOC` |
| [GO-015-SECURITY](GO-015-SECURITY.md) | Безопасность | `SEC` |
| [GO-016-TESTING](GO-016-TESTING.md) | Тестирование с testify | `TEST` |
| [GO-017-MODERNIZE](GO-017-MODERNIZE.md) | Модернизация кода | `MOD` |

## Структура документа правил

1. Заголовок и одна фраза о назначении.
2. **Навигация** — оглавление документа.
3. **Таблица правил** — `PREFIX-NNN` + краткая императивная формулировка.
4. **Принцип** — зачем нужен набор правил.
5. **Подробное описание** — раскрытие каждого правила.
6. **Пример** — код с привязкой к ID и парой ❌/✅.
7. **Исключения и граничные случаи**.
8. **Версионирование**.
