package codegen

import (
	"bytes"
	"strings"
	"testing"

	"github.com/jrandolf/envgen/internal/model"
)

func TestGenerateTypeScriptBasic(t *testing.T) {
	schema := &model.SchemaFile{
		Vars: []model.VarDef{
			{Name: "DATABASE_URL", Type: model.TypeURL, Required: true, Sensitive: true, Docs: "PostgreSQL connection string"},
			{Name: "PORT", Type: model.TypePort, Required: true, HasDefault: true, Default: "3000", Docs: "HTTP port"},
			{Name: "ANTHROPIC_API_KEY", Type: model.TypeString, Sensitive: true, Docs: "API key"},
			{Name: "NATS_REPLICAS", Type: model.TypeNumber, Required: true, Docs: "Stream replicas"},
		},
	}

	var buf bytes.Buffer
	if err := GenerateTypeScript(&buf, schema); err != nil {
		t.Fatal(err)
	}

	out := buf.String()

	// Check Secret class is generated.
	if !strings.Contains(out, "export class Secret") {
		t.Error("missing Secret class")
	}
	if !strings.Contains(out, "#value") {
		t.Error("missing private #value field in Secret")
	}
	if !strings.Contains(out, "expose(): string") {
		t.Error("missing expose() method in Secret")
	}
	if !strings.Contains(out, `"[REDACTED]"`) {
		t.Error("missing REDACTED in Secret")
	}
	if !strings.Contains(out, "nodejs.util.inspect.custom") {
		t.Error("missing inspect custom symbol in Secret")
	}

	// Check interface.
	if !strings.Contains(out, "export interface Env") {
		t.Error("missing Env interface")
	}

	// Check sensitive field uses Secret type.
	if !strings.Contains(out, "databaseUrl: Secret;") {
		t.Errorf("missing databaseUrl Secret field:\n%s", out)
	}
	if !strings.Contains(out, "port: number;") {
		t.Error("missing port field")
	}
	if !strings.Contains(out, "anthropicApiKey: Secret | undefined;") {
		t.Errorf("missing optional anthropicApiKey Secret field:\n%s", out)
	}
	if !strings.Contains(out, "natsReplicas: number;") {
		t.Error("missing natsReplicas field")
	}

	// Check loadEnv wraps sensitive values.
	if !strings.Contains(out, "new Secret(databaseUrl)") {
		t.Errorf("missing Secret wrapping in loadEnv:\n%s", out)
	}

	// Check loadEnv function.
	if !strings.Contains(out, "export const loadEnv = (): Env =>") {
		t.Error("missing loadEnv function")
	}

	// Check required validation.
	if !strings.Contains(out, `"DATABASE_URL is required"`) {
		t.Error("missing DATABASE_URL required check")
	}

	// Check default fallback.
	if !strings.Contains(out, `|| "3000"`) {
		t.Error("missing PORT default fallback")
	}

	// Check parseInt for numeric types.
	if !strings.Contains(out, "parseInt(portRaw, 10)") {
		t.Error("missing PORT parseInt")
	}

	// Check error throw.
	if !strings.Contains(out, "throw new Error") {
		t.Error("missing error throw")
	}
}

func TestGenerateTypeScriptNoSensitive(t *testing.T) {
	schema := &model.SchemaFile{
		Vars: []model.VarDef{
			{Name: "PORT", Type: model.TypePort, Required: true, HasDefault: true, Default: "3000"},
		},
	}

	var buf bytes.Buffer
	if err := GenerateTypeScript(&buf, schema); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if strings.Contains(out, "class Secret") {
		t.Error("should not emit Secret class when no sensitive vars")
	}
}

func TestEnvToTSName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"DATABASE_URL", "databaseUrl"},
		{"PORT", "port"},
		{"NATS_URL", "natsUrl"},
		{"ANTHROPIC_API_KEY", "anthropicApiKey"},
		{"AUTH_JWKS_URL", "authJwksUrl"},
		{"MAX_CONCURRENT_RESEARCH", "maxConcurrentResearch"},
		{"NEXT_PUBLIC_API_URL", "nextPublicApiUrl"},
		{"NODE_ENV", "nodeEnv"},
		{"OTEL_SERVICE_NAME", "otelServiceName"},
	}
	for _, tt := range tests {
		got := envToTSName(tt.input)
		if got != tt.want {
			t.Errorf("envToTSName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestGenerateTypeScriptEnum(t *testing.T) {
	schema := &model.SchemaFile{
		Vars: []model.VarDef{
			{Name: "NODE_ENV", Type: model.TypeEnum, EnumValues: []string{"development", "production", "test"}, Required: true, Docs: "Node environment"},
		},
	}

	var buf bytes.Buffer
	if err := GenerateTypeScript(&buf, schema); err != nil {
		t.Fatal(err)
	}

	out := buf.String()

	// Check enum type in interface.
	if !strings.Contains(out, `"development" | "production" | "test"`) {
		t.Errorf("missing enum type in interface:\n%s", out)
	}

	// Check enum validation in loadEnv.
	if !strings.Contains(out, `"development", "production", "test"`) {
		t.Errorf("missing enum validation:\n%s", out)
	}
}

func TestGenerateTypeScriptOptionalNumeric(t *testing.T) {
	schema := &model.SchemaFile{
		Vars: []model.VarDef{
			{Name: "TIMEOUT", Type: model.TypeNumber, Docs: "Timeout in seconds"},
		},
	}

	var buf bytes.Buffer
	if err := GenerateTypeScript(&buf, schema); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "timeout: number | undefined;") {
		t.Errorf("missing optional numeric field:\n%s", out)
	}
	if !strings.Contains(out, "let timeout: number | undefined = undefined;") {
		t.Errorf("missing optional numeric variable:\n%s", out)
	}
}

func TestGenerateTypeScriptJSDoc(t *testing.T) {
	schema := &model.SchemaFile{
		Vars: []model.VarDef{
			{Name: "API_URL", Type: model.TypeURL, Required: true, Docs: "Backend API URL", DocsURL: "https://example.com/docs"},
		},
	}

	var buf bytes.Buffer
	if err := GenerateTypeScript(&buf, schema); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "Backend API URL") || !strings.Contains(out, "{@link https://example.com/docs}") {
		t.Errorf("missing JSDoc link:\n%s", out)
	}
}
