# Go: Работа с базами данных
Практики работы с БД (PostgreSQL, MySQL/MariaDB, SQLite) через `database/sql`/`sqlx`/`pgx`: безопасность, контекст, ресурсы, транзакции и пул соединений.

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
| DB-001 | Использовать parameterized queries (`$1`, `?`); никогда не конкатенировать SQL. |
| DB-002 | Все обращения к БД делать через `*Context`-методы (`QueryContext`, `ExecContext`, `GetContext`). |
| DB-003 | Ставить `defer rows.Close()` сразу после `QueryContext`. |
| DB-004 | Проверять `rows.Err()` после итерации по строкам. |
| DB-005 | Обрабатывать `sql.ErrNoRows` через `errors.Is` и переводить в доменную ошибку. |
| DB-006 | Оборачивать multi-statement операции в транзакцию (`BeginTxx`/`Commit`/`Rollback`). |
| DB-007 | Использовать `SELECT ... FOR UPDATE` при чтении строк для последующего изменения. |
| DB-008 | Задавать уровень изоляции транзакции, когда default недостаточен. |
| DB-009 | Обрабатывать nullable-колонки через pointer-поля или `sql.NullXxx`. |
| DB-010 | Настраивать пул: `SetMaxOpenConns`, `SetMaxIdleConns`, `SetConnMaxLifetime`. |
| DB-011 | Управлять миграциями внешним инструментом (golang-migrate/Flyway), а не вручную. |
| DB-012 | Не использовать ORM, скрывающие SQL — работать через `sqlx`/`pgx` напрямую. |

---

## Принцип
Работа с БД в Go строится вокруг четырёх свойств: безопасность (параметризация против SQL-инъекций), управляемость (context для таймаутов и отмены), отсутствие утечек (закрытие rows и настроенный пул) и целостность (транзакции для многошаговых изменений). Ошибки здесь дороги: строковая конкатенация открывает инъекцию, забытый `rows.Close()` исчерпывает пул соединений, а отсутствие транзакции оставляет данные в неконсистентном состоянии.

| Понятие | Описание |
|---------|----------|
| `Parameterized query` | Запрос с placeholders, где значения передаются отдельно от SQL. |
| `Connection leak` | Незакрытые `rows`/соединения, исчерпывающие пул. |
| `Isolation level` | Уровень изоляции транзакции, определяющий видимость параллельных изменений. |

---

## Подробное описание

**[DB-001]** Значения передавайте только через placeholders (`$1` для PostgreSQL, `?` для MySQL). Конкатенация строк в SQL — прямой путь к инъекции.

**[DB-002]** Все операции — через `*Context`-методы, чтобы работали таймауты и отмена запроса вместе с request context.

**[DB-003]** Сразу после `QueryContext` ставьте `defer rows.Close()`. Иначе на error-пути соединение не вернётся в пул.

**[DB-004]** После цикла `rows.Next()` проверяйте `rows.Err()` — итерация может прерваться из-за ошибки, которую `Next()` не вернёт.

**[DB-005]** «Не найдено» — это `sql.ErrNoRows`; проверяйте через `errors.Is` и переводите в доменную ошибку (`ErrUserNotFound`), отделяя от инфраструктурных сбоев.

**[DB-006]** Несколько взаимосвязанных изменений оборачивайте в транзакцию: `BeginTxx` → операции → `Commit`, с `Rollback` на ошибке. Иначе возможна частично применённая запись.

**[DB-007]** Если строка читается, чтобы затем быть изменённой, используйте `SELECT ... FOR UPDATE` — это предотвращает гонки между чтением и записью.

**[DB-008]** Когда дефолтный уровень изоляции недостаточен (например, нужен `Serializable`), задавайте его явно через `sql.TxOptions`.

**[DB-009]** Nullable-колонки маппьте на pointer-поля (`*string`) или `sql.NullString`/`sql.NullInt64` — иначе `Scan` паникует на NULL.

**[DB-010]** Настраивайте пул соединений (`SetMaxOpenConns`, `SetMaxIdleConns`, `SetConnMaxLifetime`) — дефолтные значения часто не подходят под нагрузку.

**[DB-011]** Схему меняйте через golang-migrate или Flyway с версионированными миграциями. Ручные SQL рассинхронизируют окружения.

**[DB-012]** Не используйте ORM, генерирующие непредсказуемый SQL. Работайте через `sqlx` (struct scanning) или `pgx` (оптимизации под PostgreSQL) напрямую.

> **Важно:** отсутствие транзакции в переводе средств (нарушение DB-006) может списать деньги у отправителя и не зачислить получателю — классический пример неконсистентности.

---

## Пример

```go
// [DB-001][DB-002][DB-005] параметризация, context, обработка ErrNoRows
var user User
err := db.GetContext(ctx, &user,
    "SELECT id, name, email FROM users WHERE email = $1", email)
if err != nil {
    if errors.Is(err, sql.ErrNoRows) {
        return nil, ErrUserNotFound
    }
    return nil, fmt.Errorf("querying user %s: %w", email, err)
}

// [DB-003][DB-004] закрытие rows и проверка Err
rows, err := db.QueryContext(ctx, "SELECT id, name FROM users")
if err != nil {
    return fmt.Errorf("querying users: %w", err)
}
defer rows.Close()
for rows.Next() {
    // rows.Scan(...)
}
if err := rows.Err(); err != nil {
    return fmt.Errorf("iterating users: %w", err)
}
```

```go
// ❌ Нарушение [DB-001] — конкатенация, SQL injection
query := fmt.Sprintf("SELECT * FROM users WHERE email = '%s'", email)
rows, err := db.Query(query)

// ✅ Правильно — параметризация + context
rows, err := db.QueryContext(ctx, "SELECT * FROM users WHERE email = $1", email)
defer rows.Close()
```

```go
// [DB-006] транзакция для multi-statement
tx, err := db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
if err != nil {
    return fmt.Errorf("begin tx: %w", err)
}
if _, err = tx.ExecContext(ctx,
    "UPDATE accounts SET balance = balance - $1 WHERE id = $2", amount, fromID); err != nil {
    _ = tx.Rollback()
    return fmt.Errorf("update from: %w", err)
}
if _, err = tx.ExecContext(ctx,
    "UPDATE accounts SET balance = balance + $1 WHERE id = $2", amount, toID); err != nil {
    _ = tx.Rollback()
    return fmt.Errorf("update to: %w", err)
}
if err := tx.Commit(); err != nil {
    return fmt.Errorf("commit tx: %w", err)
}
```

---

## Исключения и граничные случаи
| Ситуация | Как поступить |
|----------|---------------|
| Одиночный `INSERT/UPDATE` | Транзакция не обязательна — достаточно `ExecContext` (DB-006). |
| Bulk-вставка большого объёма | Использовать batch/`COPY` (pgx) вместо построчных запросов (DB-002). |
| Чтение без последующей модификации | `FOR UPDATE` не нужен — он лишь удерживает блокировку (DB-007). |

---

## Версионирование
| Версия | Дата | Задача | Агент | Модель | Описание изменений |
|--------|------|--------|-------|--------|--------------------|
| 1.0.0 | 2026-09-09 | Стандартизация docs/go-raw-rules | Claude Code | Opus 4.8 | Начальное создание по шаблону |
