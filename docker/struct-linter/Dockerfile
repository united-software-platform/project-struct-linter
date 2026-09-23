# Сборка статического бинаря линтера структуры и упаковка в distroless-образ.
# Схемы и эталоны вкомпилированы в бинарь через embed, поэтому финальный образ содержит только
# исполняемый файл — самодостаточный след без файлов данных на диске (Standard 20).

ARG GO_IMAGE=golang:1.24-alpine
ARG RUNTIME_IMAGE=gcr.io/distroless/static-debian12:nonroot

FROM ${GO_IMAGE} AS build
WORKDIR /src
# Слой зависимостей кэшируется отдельно от исходников.
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=dev
# CGO выключен — бинарь статический и переносим между базовыми образами.
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags "-s -w -X main.version=${VERSION}" \
    -o /out/struct-linter ./cmd/struct-linter

FROM ${RUNTIME_IMAGE}
# distroless:nonroot исполняется под непривилегированным пользователем 65532 [DOCK-012].
COPY --from=build /out/struct-linter /usr/local/bin/struct-linter
WORKDIR /work
USER nonroot:nonroot
ENTRYPOINT ["/usr/local/bin/struct-linter"]
