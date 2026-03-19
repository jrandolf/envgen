# envgen

Generate typed, validated config code from `.env.schema` files.

Reads a `.env.schema` file and emits a typed config module for Go, Python, or TypeScript. The generated code validates all environment variables at startup, applies schema defaults, and wraps sensitive values in redaction-safe types.

## Install

```bash
go install github.com/jrandolf/envgen@latest
```

## Usage

```bash
envgen -lang=go  -schema=.env.schema -out=config/config.go -package=config
envgen -lang=py  -schema=.env.schema -out=src/config.py
envgen -lang=ts  -schema=.env.schema -out=lib/env.ts
```

## Skipping validation

Set `SKIP_ENV_VALIDATION` to any non-empty value to suppress validation errors at startup.
This is useful in test environments where not all variables are populated:

    SKIP_ENV_VALIDATION=true node my-app.js

When set, missing required variables receive empty string / zero values and the process
continues. Type and enum checks are still collected but not raised.

## Schema format

```bash
# @defaultSensitive=false
# @defaultRequired=true
# ---

# @sensitive @type=url
# @docs("PostgreSQL connection string")
DATABASE_URL=

# @type=port
# @docs("HTTP server port")
PORT=3000

# @optional @sensitive
# @docs("API key for external service")
API_KEY=

# @type=number
# @docs("Max concurrent requests")
MAX_CONCURRENCY=10

# @type=enum(development, production, test)
# @docs("Runtime environment")
NODE_ENV=
```

### Annotations

| Annotation | Description |
|---|---|
| `@type=port` | Integer port number |
| `@type=number` | Integer |
| `@type=url` | URL string |
| `@type=email` | Email string |
| `@type=enum(a, b, c)` | One of the listed values |
| `@sensitive` | Secret — wrapped in a redaction-safe type |
| `@optional` | Not required; may be absent |
| `@docs("description")` | Documentation comment |
| `@docs("description", https://...)` | Documentation with external link |

### Defaults

A value after `=` is a schema default. The generated `Load()`/`from_env()`/`loadEnv()` uses it as a fallback when the env var is empty. Variables with no value after `=` have no default — they fail if absent (unless `@optional`).

## Generated output

### Go (`-lang=go`)

Generates a `Config` struct with a `Load() (*Config, error)` function. Sensitive values are wrapped in a `Secret` type that returns `[REDACTED]` from `String()`, `GoString()`, and `MarshalText()`. Use `secret.Expose()` to access the underlying value.

```go
cfg, err := config.Load()
if err != nil {
    log.Fatal(err)
}
pool, err := pgxpool.New(ctx, cfg.DatabaseURL.Expose())
```

### Python (`-lang=py`)

Generates a frozen `@dataclass` with a `from_env()` classmethod. Sensitive values use `pydantic.SecretStr` — use `.get_secret_value()` to access the underlying value. `str()` returns `**********`.

```python
cfg = Config.from_env()
engine = create_engine(cfg.database_url.get_secret_value())
```

### TypeScript (`-lang=ts`)

Generates an `Env` interface and `loadEnv()` function with zero dependencies. Sensitive values are wrapped in a `Secret` class that returns `[REDACTED]` from `toString()`, `toJSON()`, and Node.js `inspect()`. Use `secret.expose()` to access the underlying value.

```typescript
const env = loadEnv();
const pool = createPool(env.databaseUrl.expose());
```

## Naming conventions

| Schema | Go | Python | TypeScript |
|---|---|---|---|
| `DATABASE_URL` | `DatabaseURL` | `database_url` | `databaseUrl` |
| `PORT` | `Port` | `port` | `port` |
| `API_KEY` | `APIKey` | `api_key` | `apiKey` |

## Type mapping

| Schema | Go | Go (sensitive) | Python | Python (sensitive) | TypeScript | TypeScript (sensitive) |
|---|---|---|---|---|---|---|
| string | `string` | `Secret` | `str` | `SecretStr` | `string` | `Secret` |
| port | `int` | — | `int` | — | `number` | — |
| number | `int` | — | `int` | — | `number` | — |
| url | `string` | `Secret` | `str` | `SecretStr` | `string` | `Secret` |
| email | `string` | — | `str` | — | `string` | — |
| enum | `string` | — | `str` | — | union literal | — |

## License

Apache 2.0 — see [LICENSE](LICENSE).
