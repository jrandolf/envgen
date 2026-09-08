# envgen

Generate typed, validated config code from `.env.schema` files.

Reads a `.env.schema` file and emits a typed config module for Go, Python, Rust, or TypeScript. The generated code validates all environment variables at startup, applies schema defaults, and wraps sensitive values in redaction-safe types.

## Install

Install a prebuilt binary with [mise](https://mise.jdx.dev/):

```bash
mise use github:jrandolf/envgen@0.1.0
envgen --version
```

Or declare the tool in `mise.toml`:

```toml
[tools]
envgen = "0.1.0"

[tool_alias]
envgen = "github:jrandolf/envgen"
```

[GitHub releases](https://github.com/jrandolf/envgen/releases) provide archives for
macOS, Linux, and Windows on amd64 (Intel/AMD) and arm64. Linux binaries are static
and work with both glibc and musl, including Alpine. No Go toolchain is required.
For manual installation, extract the archive for your platform and put `envgen`
(`envgen.exe` on Windows) on your `PATH`. Each release includes SHA-256 checksums
in `checksums.txt` and GitHub build provenance attestations.

To compile from source instead:

```bash
go install github.com/jrandolf/envgen@latest
```

## Releasing

Push a semantic version tag from `main` to publish a release:

```bash
git switch main
git pull --ff-only
git tag v0.1.0
git push origin v0.1.0
```

The release workflow runs the Go tests and vet checks, builds all six archives
with GoReleaser, verifies their checksums, and runs each binary on its native
platform. Both Linux architectures also run on Alpine. Only after every check
passes does the workflow attest and publish those same archives, then verify
installation through mise on all six platforms. Tags with a prerelease suffix
(for example, `v0.2.0-rc.1`) publish GitHub prereleases.

Pull requests and pushes to `main` run the same build and archive checks using
snapshot versions. To check the release build locally with Go and GoReleaser
2.18.1 installed:

```bash
go test ./...
go vet ./...
goreleaser check
goreleaser release --snapshot --clean
```

## Usage

```bash
envgen -lang=go  -schema=.env.schema -out=config/config.go -package=config
envgen -lang=py  -schema=.env.schema -out=src/config.py
envgen -lang=rs  -schema=.env.schema -out=src/config.rs
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

### Rust (`-lang=rs`)

Generates an `Env` struct with an `Env::from_env() -> Result<Env, String>` function. Sensitive values use [`secrecy::SecretString`](https://docs.rs/secrecy), which redacts on `Debug` and requires the `ExposeSecret` trait to read the value. Add `secrecy = "0.10"` to your `Cargo.toml` when the schema contains sensitive vars.

```rust
use config::Env;
use secrecy::ExposeSecret;

let env = Env::from_env()?;
let pool = PgPool::connect(env.database_url.expose_secret()).await?;
```

### TypeScript (`-lang=ts`)

Generates an `Env` interface and `loadEnv()` function with zero dependencies. Sensitive values are wrapped in a `Secret` class that returns `[REDACTED]` from `toString()`, `toJSON()`, and Node.js `inspect()`. Use `secret.expose()` to access the underlying value.

```typescript
const env = loadEnv();
const pool = createPool(env.databaseUrl.expose());
```

## Naming conventions

| Schema | Go | Python | Rust | TypeScript |
|---|---|---|---|---|
| `DATABASE_URL` | `DatabaseURL` | `database_url` | `database_url` | `databaseUrl` |
| `PORT` | `Port` | `port` | `port` | `port` |
| `API_KEY` | `APIKey` | `api_key` | `api_key` | `apiKey` |

## Type mapping

| Schema | Go | Go (sensitive) | Python | Python (sensitive) | Rust | Rust (sensitive) | TypeScript | TypeScript (sensitive) |
|---|---|---|---|---|---|---|---|---|
| string | `string` | `Secret` | `str` | `SecretStr` | `String` | `Secret` | `string` | `Secret` |
| port | `int` | — | `int` | — | `i64` | — | `number` | — |
| number | `int` | — | `int` | — | `i64` | — | `number` | — |
| url | `string` | `Secret` | `str` | `SecretStr` | `String` | `Secret` | `string` | `Secret` |
| email | `string` | — | `str` | — | `String` | — | `string` | — |
| enum | `string` | — | `str` | — | `String` | — | union literal | — |

## License

Apache 2.0 — see [LICENSE](LICENSE).
