package codegen

import (
	"bytes"
	"strings"
	"testing"

	"github.com/jrandolf/envgen/internal/model"
)

func TestGenerateGoBasic(t *testing.T) {
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
	if err := GenerateGo(&buf, schema, "config"); err != nil {
		t.Fatal(err)
	}

	out := buf.String()

	// Check package declaration.
	if !strings.Contains(out, "package config") {
		t.Error("missing package declaration")
	}

	// Check Secret type is generated.
	if !strings.Contains(out, "type Secret struct") {
		t.Error("missing Secret type")
	}

	// Check struct fields.
	if !strings.Contains(out, "DatabaseURL Secret") {
		t.Errorf("missing DatabaseURL Secret field")
	}
	if !strings.Contains(out, "Port int") {
		t.Errorf("missing Port int field")
	}
	if !strings.Contains(out, "AnthropicAPIKey Secret") {
		t.Errorf("missing AnthropicAPIKey Secret field")
	}
	if !strings.Contains(out, "NATSReplicas int") {
		t.Errorf("missing NATSReplicas int field")
	}

	// Check load function and Cfg var.
	if !strings.Contains(out, "func load() (*Config, error)") {
		t.Error("missing load function")
	}
	if !strings.Contains(out, "var Cfg *Config") {
		t.Error("missing Cfg var")
	}
	if !strings.Contains(out, "func init()") {
		t.Error("missing init function")
	}

	// Check default fallback.
	if !strings.Contains(out, `port = "50052"`) {
		t.Error("missing PORT default fallback")
	}

	// Check required validation.
	if !strings.Contains(out, `"DATABASE_URL is required"`) {
		t.Error("missing DATABASE_URL required check")
	}
}

func TestEnvToGoName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"DATABASE_URL", "DatabaseURL"},
		{"PORT", "Port"},
		{"NATS_NKEY_SEED_PATH", "NATSNkeySeedPath"},
		{"ANTHROPIC_API_KEY", "AnthropicAPIKey"},
		{"AUTH_JWKS_URL", "AuthJWKSURL"},
		{"OUTBOX_RESEARCHER_URL", "OutboxResearcherURL"},
		{"OTEL_EXPORTER_OTLP_ENDPOINT", "OTELExporterOtlpEndpoint"},
		{"REDIS_URL", "RedisURL"},
		{"OUTBOX_APP_URL", "OutboxAppURL"},
		{"MAX_CONCURRENT_RESEARCH", "MaxConcurrentResearch"},
		{"EMBEDDING_DIMENSION", "EmbeddingDimension"},
		{"OUTBOX_CREDENTIALS_API_KEY", "OutboxCredentialsAPIKey"},
		{"GOOGLE_OAUTH_CLIENT_ID", "GoogleOauthClientID"},
		{"OUTBOX_RESEARCHER_TIMEOUT_SECONDS", "OutboxResearcherTimeoutSeconds"},
		{"OAUTH_STATE_SECRET", "OauthStateSecret"},
		{"API_KEYS", "APIKeys"},
		{"OTEL_SERVICE_NAME", "OTELServiceName"},
	}
	for _, tt := range tests {
		got := envToGoName(tt.input)
		if got != tt.want {
			t.Errorf("envToGoName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestEnvToLocalName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"DATABASE_URL", "databaseURL"},
		{"PORT", "port"},
		{"NATS_URL", "natsURL"},
		{"ANTHROPIC_API_KEY", "anthropicAPIKey"},
		{"AUTH_JWKS_URL", "authJWKSURL"},
		{"API_KEYS", "apiKeys"},
		{"REDIS_URL", "redisURL"},
	}
	for _, tt := range tests {
		got := envToLocalName(tt.input)
		if got != tt.want {
			t.Errorf("envToLocalName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestGenerateGoOptionalString(t *testing.T) {
	schema := &model.SchemaFile{
		Vars: []model.VarDef{
			{Name: "ANTHROPIC_BASE_URL", Type: model.TypeURL, Required: false, Docs: "API base URL override"},
		},
	}

	var buf bytes.Buffer
	if err := GenerateGo(&buf, schema, "config"); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	// Optional non-sensitive URL should be a plain string.
	if !strings.Contains(out, "AnthropicBaseURL string") {
		t.Errorf("expected AnthropicBaseURL string, got:\n%s", out)
	}
}

func TestGenerateGoEnum(t *testing.T) {
	schema := &model.SchemaFile{
		Vars: []model.VarDef{
			{Name: "NODE_ENV", Type: model.TypeEnum, EnumValues: []string{"development", "production", "test"}, Required: true, Docs: "Node environment"},
		},
	}

	var buf bytes.Buffer
	if err := GenerateGo(&buf, schema, "config"); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, `"development", "production", "test"`) {
		t.Errorf("missing enum validation:\n%s", out)
	}
}

func TestGenerateGoSkipValidation(t *testing.T) {
	schema := &model.SchemaFile{
		Vars: []model.VarDef{
			{Name: "REQUIRED_VAR", Type: model.TypeString, Required: true, Docs: "Required test variable"},
		},
	}

	var buf bytes.Buffer
	if err := GenerateGo(&buf, schema, "config"); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, `os.Getenv("SKIP_ENV_VALIDATION") == ""`) {
		t.Errorf("expected SKIP_ENV_VALIDATION check in error gate:\n%s", out)
	}
}
