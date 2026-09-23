# Go: Работа с context.Context
Правила создания, распространения, отмены и таймаутов `context.Context`, а также корректной передачи request-scoped метаданных.

---

## Навигация
- [Таблица правил](#таблица-правил)
- [Принцип](#принцип)
- [Подробное описание](#подробное-описание)
- [Пример](#пример)
- [Исключения и граничные случаи](#исключения-и-граничные-случаи)
- [Версионирование](#версионирование)

---

## Таблица правил
| ID | Правило |
|----|---------|
| CTX-001 | Передавать `context.Context` первым параметром через весь request lifecycle. |
| CTX-002 | Не создавать `context.Background()` в середине request path — это рвёт цепочку отмены. |
| CTX-003 | Не хранить context в полях struct — передавать явно через параметры. |
| CTX-004 | Не передавать `nil` context; при неизвестности использовать `context.TODO()`. |
| CTX-005 | `context.Background()` использовать только на верхнем уровне (main, init, tests). |
| CTX-006 | В HTTP handler использовать `r.Context()`, в gRPC — `stream.Context()`. |
| CTX-007 | Устанавливать `context.WithTimeout`/`WithDeadline` для всех внешних вызовов (DB, HTTP, gRPC). |
| CTX-008 | Хранить в context только request-scoped метаданные, не аргументы функций. |
| CTX-009 | Для ключей context использовать unexported тип, а не строку (защита от коллизий). |
| CTX-010 | Вызывать `cancel()` из `WithCancel`/`WithTimeout` через `defer`. |
| CTX-011 | В долгоживущих горутинах слушать `ctx.Done()` для остановки. |

---

## Принцип
`context.Context` — сквозной механизм отмены, дедлайнов и переноса request-scoped метаданных. Он должен непрерывно течь по цепочке вызовов от точки входа (HTTP/gRPC) до внешних вызовов. Любое создание нового `context.Background()` в середине пути рвёт цепочку: клиент отменил запрос, а сервер продолжает работать и течёт ресурсами. Значения в context предназначены для метаданных (request ID, user ID, trace ID), а не для передачи параметров функций.

| Понятие | Описание |
|---------|----------|
| `Propagation` | Непрерывная передача одного context по всей цепочке вызовов. |
| `Cancellation chain` | Дерево производных context'ов, отменяемых вместе с родителем. |
| `Request-scoped metadata` | Данные, привязанные к запросу (request/trace/user ID), но не бизнес-аргументы. |

---

## Подробное описание

**[CTX-001]** `context.Context` — всегда первый параметр (`ctx context.Context`), передаваемый через всю цепочку: HTTP handler → service → repository → внешние вызовы.

**[CTX-002]** Не создавайте новый `context.Background()` внутри request path — это разрывает цепочку отмены. Передавайте полученный `ctx` дальше.

**[CTX-003]** Не храните context в полях структуры: он привязан к запросу, а структура живёт дольше. Передавайте явно параметром.

**[CTX-004]** Не передавайте `nil` context — это паникует у корректных вызовов. Если непонятно, какой context использовать, ставьте `context.TODO()`.

**[CTX-005]** `context.Background()` — только в корне: `main`, `init`, тесты. Оттуда он растекается вниз производными context'ами.

**[CTX-006]** В HTTP handler источник — `r.Context()`, в gRPC — `stream.Context()`/`req` context. Не подменяйте их `context.Background()`.

**[CTX-007]** Для каждого внешнего вызова (БД, HTTP, gRPC) задавайте `WithTimeout`/`WithDeadline`, иначе вызов может зависнуть навсегда.

**[CTX-008]** В context кладите только request-scoped метаданные (request ID, user ID, trace ID). Бизнес-аргументы передавайте явными параметрами.

**[CTX-009]** Ключи для `WithValue` объявляйте как unexported тип (`type contextKey string`), а не как голую строку — иначе возможны коллизии между пакетами.

**[CTX-010]** `WithCancel`/`WithTimeout`/`WithDeadline` возвращают `cancel` — вызывайте его через `defer cancel()`, иначе течёт контекст и таймер.

**[CTX-011]** Долгоживущие горутины должны слушать `ctx.Done()` в `select` и завершаться при отмене, возвращая `ctx.Err()`.

> **Важно:** Значения context — для метаданных, не для параметров (CTX-008). Использование `ctx.Value` вместо явных аргументов делает контракт функции невидимым и небезопасным по типам.

---

## Пример

```go
// [CTX-006] handler берёт context из запроса, [CTX-007] задаёт timeout
func handler(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
    defer cancel() // [CTX-010]
    result := process(ctx)
    _, _ = w.Write([]byte(result))
}

// [CTX-009] unexported тип ключа для метаданных
type contextKey string

const requestIDKey contextKey = "request_id"
```

```go
// ❌ Нарушение [CTX-002] — новый Background рвёт цепочку отмены
func (s *OrderService) Create(ctx context.Context, o Order) error {
    return s.db.ExecContext(context.Background(), "INSERT INTO orders ...", o.ID)
}

// ✅ Правильно — передаём полученный ctx дальше
func (s *OrderService) Create(ctx context.Context, o Order) error {
    return s.db.ExecContext(ctx, "INSERT INTO orders ...", o.ID)
}
```

```go
// [CTX-011] горутина завершается по отмене
func worker(ctx context.Context, ch <-chan Task) {
    for {
        select {
        case <-ctx.Done():
            return
        case task, ok := <-ch:
            if !ok {
                return
            }
            process(task)
        }
    }
}
```

---

## Исключения и граничные случаи
| Ситуация | Как поступить |
|----------|---------------|
| Операция должна пережить отмену родителя (например, фоновый flush) | Использовать `context.WithoutCancel(parent)` (Go 1.21+), не `Background()` (CTX-002). |
| Неизвестен корректный context при рефакторинге | Временно `context.TODO()` с TODO-комментарием (CTX-004). |
| Верхний уровень приложения | `context.Background()` допустим только здесь (CTX-005). |

---

## Версионирование
| Версия | Дата | Задача | Агент | Модель | Описание изменений |
|--------|------|--------|-------|--------|--------------------|
| 1.0.0 | 2026-09-09 | Стандартизация docs/go-raw-rules | Claude Code | Opus 4.8 | Начальное создание по шаблону |
