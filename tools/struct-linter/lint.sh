#!/bin/sh
# Обёртка запуска линтера структуры в контейнере: монтирует текущий каталог только для чтения
# и пробрасывает аргументы в бинарь. Линтер читает только имена и пути (:ro), содержимое файлов
# проекта не читается.
#
# Команда docker run собирается из явных аргументов, а не из строки, интерполирующей внешние
# данные (SEC-003). Код возврата контейнера пробрасывается через exec без изменений: 0 — ok,
# 1 — нарушения, 2 — ошибка вызова.
#
# Пример:
#   tools/struct-linter/lint.sh validate --template auto
#   tools/struct-linter/lint.sh list-templates

set -eu

IMAGE="${STRUCT_LINTER_IMAGE:-struct-linter:local}"

exec docker run --rm -v "$PWD:/work:ro" "$IMAGE" "$@"
