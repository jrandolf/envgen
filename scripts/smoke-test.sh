#!/bin/sh
set -eu

binary=$1
version=$2
test "$("$binary" --version)" = "envgen $version"

workdir=$(mktemp -d)
trap 'rm -rf "$workdir"' EXIT
cat > "$workdir/.env.schema" <<'SCHEMA'
# @type=port
PORT=3000
# @optional
API_KEY=
SCHEMA

for lang in go py rs ts; do
  "$binary" -lang="$lang" -schema="$workdir/.env.schema" -out="$workdir/config.$lang"
  test -s "$workdir/config.$lang"
done

grep -q 'type Config struct' "$workdir/config.go"
grep -q 'class Config' "$workdir/config.py"
grep -q 'pub struct Env' "$workdir/config.rs"
grep -q 'interface Env' "$workdir/config.ts"
printf 'envgen %s: version and all four generators passed\n' "$version"

