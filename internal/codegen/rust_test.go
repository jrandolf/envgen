package codegen

import (
	"bytes"
	"strings"
	"testing"

	"github.com/jrandolf/envgen/internal/model"
)

func TestGenerateRustBasic(t *testing.T) {
	schema := &model.SchemaFile{
		DefaultRequired:  true,
		DefaultSensitive: false,
		Vars: []model.VarDef{
			{Name: "DATABASE_URL", Type: model.TypeURL, Required: true, Sensitive: true, Docs: "PostgreSQL connection string"},
			{Name: "PORT", Type: model.TypePort, Required: true, HasDefault: true, Default: "50052", Docs: "HTTP port"},
			{Name: "ANTHROPIC_API_KEY", Type: model.TypeString, Sensitive: true, Docs: "API key"},
			{Name: "NATS_REPLICAS", Type: model.TypeNumber, Required: true, Docs: "Stream replicas"},
		},
	}

	var buf bytes.Buffer
	if err := GenerateRust(&buf, schema); err != nil {
		t.Fatal(err)
	}

	out := buf.String()

	// Check secrecy crate is imported.
	if !strings.Contains(out, "use secrecy::SecretString;") {
		t.Error("missing secrecy::SecretString import")
	}

	// Check struct fields.
	if !strings.Contains(out, "pub database_url: SecretString") {
		t.Errorf("missing database_url SecretString field")
	}
	if !strings.Contains(out, "pub port: i64") {
		t.Errorf("missing port i64 field")
	}
	if !strings.Contains(out, "pub anthropic_api_key: Option<SecretString>") {
		t.Errorf("missing anthropic_api_key Option<SecretString> field")
	}
	if !strings.Contains(out, "pub nats_replicas: i64") {
		t.Errorf("missing nats_replicas i64 field")
	}

	// Check load function and CONFIG static.
	if !strings.Contains(out, "pub fn load() -> Result<Config, String>") {
		t.Error("missing load function")
	}
	if !strings.Contains(out, "pub static CONFIG: LazyLock<Config>") {
		t.Error("missing CONFIG static")
	}

	// Check default fallback.
	if !strings.Contains(out, `"50052".to_string()`) {
		t.Error("missing PORT default fallback")
	}

	// Check required validation.
	if !strings.Contains(out, `"DATABASE_URL is required"`) {
		t.Error("missing DATABASE_URL required check")
	}

	// Check SecretString wrapping in constructor.
	if !strings.Contains(out, "SecretString::from(database_url)") {
		t.Error("missing SecretString wrapping for database_url")
	}
}

func TestGenerateRustNoSensitive(t *testing.T) {
	schema := &model.SchemaFile{
		Vars: []model.VarDef{
			{Name: "PORT", Type: model.TypePort, Required: true, HasDefault: true, Default: "3000"},
		},
	}

	var buf bytes.Buffer
	if err := GenerateRust(&buf, schema); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if strings.Contains(out, "use secrecy::") {
		t.Error("should not import secrecy when no sensitive vars")
	}
	if strings.Contains(out, "SecretString") {
		t.Error("should not reference SecretString when no sensitive vars")
	}
}

func TestEnvToRustName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"DATABASE_URL", "database_url"},
		{"PORT", "port"},
		{"NATS_URL", "nats_url"},
		{"ANTHROPIC_API_KEY", "anthropic_api_key"},
		{"AUTH_JWKS_URL", "auth_jwks_url"},
		{"MAX_CONCURRENT_RESEARCH", "max_concurrent_research"},
		{"NODE_ENV", "node_env"},
		{"OTEL_SERVICE_NAME", "otel_service_name"},
	}
	for _, tt := range tests {
		got := envToRustName(tt.input)
		if got != tt.want {
			t.Errorf("envToRustName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestGenerateRustEnum(t *testing.T) {
	schema := &model.SchemaFile{
		Vars: []model.VarDef{
			{Name: "NODE_ENV", Type: model.TypeEnum, EnumValues: []string{"development", "production", "test"}, Required: true, Docs: "Node environment"},
		},
	}

	var buf bytes.Buffer
	if err := GenerateRust(&buf, schema); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, `"development" | "production" | "test"`) {
		t.Errorf("missing enum match arms:\n%s", out)
	}
	if !strings.Contains(out, `"NODE_ENV is required"`) {
		t.Errorf("missing required check for enum:\n%s", out)
	}
}

func TestGenerateRustOptionalString(t *testing.T) {
	schema := &model.SchemaFile{
		Vars: []model.VarDef{
			{Name: "ANTHROPIC_BASE_URL", Type: model.TypeURL, Required: false, Docs: "API base URL override"},
		},
	}

	var buf bytes.Buffer
	if err := GenerateRust(&buf, schema); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "pub anthropic_base_url: Option<String>") {
		t.Errorf("expected Option<String>, got:\n%s", out)
	}
}

func TestGenerateRustOptionalNumeric(t *testing.T) {
	schema := &model.SchemaFile{
		Vars: []model.VarDef{
			{Name: "TIMEOUT", Type: model.TypeNumber, Docs: "Timeout in seconds"},
		},
	}

	var buf bytes.Buffer
	if err := GenerateRust(&buf, schema); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "pub timeout: Option<i64>") {
		t.Errorf("missing optional numeric field:\n%s", out)
	}
	if !strings.Contains(out, "let timeout: Option<i64>") {
		t.Errorf("missing optional numeric variable:\n%s", out)
	}
}

func TestGenerateRustSkipValidation(t *testing.T) {
	schema := &model.SchemaFile{
		Vars: []model.VarDef{
			{Name: "REQUIRED_VAR", Type: model.TypeString, Required: true, Docs: "Required test variable"},
		},
	}

	var buf bytes.Buffer
	if err := GenerateRust(&buf, schema); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, `env::var("SKIP_ENV_VALIDATION")`) {
		t.Errorf("expected SKIP_ENV_VALIDATION check in error gate:\n%s", out)
	}
}

func TestGenerateRustDocComments(t *testing.T) {
	schema := &model.SchemaFile{
		Vars: []model.VarDef{
			{Name: "API_URL", Type: model.TypeURL, Required: true, Docs: "Backend API URL"},
		},
	}

	var buf bytes.Buffer
	if err := GenerateRust(&buf, schema); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "/// Backend API URL") {
		t.Errorf("missing doc comment:\n%s", out)
	}
}
