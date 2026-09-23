# Go: gRPC
Серверы и клиенты gRPC: status codes, deadlines, переиспользование соединений, health check, interceptors, TLS/mTLS и graceful shutdown.

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
| GRPC-001 | Возвращать ошибки через `status.Errorf` с конкретными `codes.*`, а не raw error. |
| GRPC-002 | Устанавливать deadline (`context.WithTimeout`) на каждый client call. |
| GRPC-003 | Переиспользовать одно соединение для множества RPC (HTTP/2 multiplexing). |
| GRPC-004 | Регистрировать `grpc_health_v1` health check service для Kubernetes probes. |
| GRPC-005 | Отключать reflection в production. |
| GRPC-006 | Выносить cross-cutting concerns (logging, auth, recovery) в interceptors. |
| GRPC-007 | Использовать TLS/mTLS через `grpc.WithTransportCredentials`. |
| GRPC-008 | Реализовывать graceful shutdown через `GracefulStop` с fallback к `Stop()`. |
| GRPC-009 | Передавать metadata (auth, trace ID) через `metadata.NewOutgoingContext`. |
| GRPC-010 | Проверять отмену context в длительных операциях и стримах. |
| GRPC-011 | Использовать сообщения (message) как аргументы RPC, а не голые типы — ради расширяемости. |

---

## Принцип
gRPC — это контракт поверх HTTP/2. Его надёжность держится на явных status codes (чтобы клиент мог принять решение о retry), дедлайнах (чтобы медленный upstream не вешал горутины), переиспользовании соединений (multiplexing вместо handshake на каждый вызов) и вынесении сквозной логики в interceptors. Health check и graceful shutdown делают сервис управляемым в Kubernetes, а TLS — безопасным в проде.

| Понятие | Описание |
|---------|----------|
| `Status code` | Типизированный код ошибки gRPC (`codes.NotFound`, `codes.Internal`), понятный клиенту. |
| `Interceptor` | Middleware для unary/stream RPC: логирование, auth, recovery. |
| `Multiplexing` | Много RPC поверх одного HTTP/2-соединения. |

---

## Подробное описание

**[GRPC-001]** Все ошибки возвращайте через `status.Errorf(codes.X, ...)`. Raw `fmt.Errorf` приходит клиенту как `codes.Unknown`, и клиент не может отличить NotFound от Internal и принять решение о retry.

**[GRPC-002]** На каждый client call ставьте deadline через `context.WithTimeout`. Без него медленный upstream держит горутину и ресурсы бесконечно.

**[GRPC-003]** Создавайте `grpc.NewClient` один раз на приложение и переиспользуйте соединение для всех RPC. Соединение на запрос — лишние TCP/TLS handshakes.

**[GRPC-004]** Регистрируйте `grpc_health_v1` health service, чтобы Kubernetes readiness/liveness probes видели готовность и не убивали pod преждевременно.

**[GRPC-005]** Reflection раскрывает всю поверхность API — отключайте его в production (не вызывайте `reflection.Register`).

**[GRPC-006]** Логирование, аутентификацию и recovery от паник выносите в interceptors (`ChainUnaryInterceptor`), а не дублируйте в каждом методе.

**[GRPC-007]** В production включайте TLS (или mTLS с проверкой клиентского сертификата) через `grpc.WithTransportCredentials`.

**[GRPC-008]** Останавливайте сервер через `GracefulStop`, дожидаясь завершения активных RPC, с таймаутом и fallback к `Stop()` для форсированной остановки.

**[GRPC-009]** Auth-токены и trace ID передавайте через `metadata.NewOutgoingContext`, а не как поля сообщения.

**[GRPC-010]** В длительных операциях и стримах слушайте `ctx.Done()`/`stream.Context()` и прекращайте работу при отмене вызывающей стороной.

**[GRPC-011]** Аргументы RPC — всегда protobuf-message, а не голые скаляры: это позволяет добавлять поля без breaking change.

> **Важно:** использование `codes.Internal` (или raw error) для всех ошибок ломает клиентскую retry-логику (нарушение GRPC-001): клиент повторяет невосстановимые ошибки и не повторяет восстановимые.

---

## Пример

```go
// [GRPC-001] конкретные status codes
func (s *myService) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
    user, err := s.db.GetUser(ctx, req.Id)
    switch {
    case errors.Is(err, ErrNotFound):
        return nil, status.Errorf(codes.NotFound, "user %q not found", req.Id)
    case errors.Is(err, ErrPermission):
        return nil, status.Errorf(codes.PermissionDenied, "access denied for %q", req.Id)
    case err != nil:
        return nil, status.Errorf(codes.Internal, "database error: %v", err)
    }
    return user, nil
}
```

```go
// [GRPC-002][GRPC-003][GRPC-009] deadline, переиспользование conn, metadata
func GetUser(ctx context.Context, conn *grpc.ClientConn, id, token string) (*pb.User, error) {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs("authorization", "Bearer "+token))
    return pb.NewMyServiceClient(conn).GetUser(ctx, &pb.GetUserRequest{Id: id})
}
```

```go
// [GRPC-008] graceful shutdown с fallback
stopped := make(chan struct{})
go func() { srv.GracefulStop(); close(stopped) }()
select {
case <-stopped:
case <-time.After(15 * time.Second):
    srv.Stop() // форсированная остановка по таймауту
}
```

---

## Исключения и граничные случаи
| Ситуация | Как поступить |
|----------|---------------|
| Локальная разработка/отладка | Reflection допустимо включать вне production (GRPC-005). |
| Внутренний доверенный периметр без TLS-терминации | TLS всё равно предпочтителен; отключение — осознанное исключение (GRPC-007). |
| Стриминговый RPC | Deadlines заменяются на слушание `stream.Context().Done()` (GRPC-002/GRPC-010). |

---

## Версионирование
| Версия | Дата | Задача | Агент | Модель | Описание изменений |
|--------|------|--------|-------|--------|--------------------|
| 1.0.0 | 2026-09-09 | Стандартизация docs/go-raw-rules | Claude Code | Opus 4.8 | Начальное создание по шаблону |
