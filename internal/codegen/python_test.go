package codegen

import (
	"bytes"
	"strings"
	"testing"

	"github.com/jrandolf/envgen/internal/model"
)

func TestGeneratePythonBasic(t *testing.T) {
	schema := &model.SchemaFile{
		DefaultRequired:  true,
		DefaultSensitive: false,
		Vars: []model.VarDef{
			{Name: "DATABASE_URL", Type: model.TypeURL, Required: true, Sensitive: true, Docs: "PostgreSQL connection string"},
			{Name: "PORT", Type: model.TypePort, Required: true, HasDefault: true, Default: "8080", Docs: "HTTP port"},
			{Name: "MAX_CONCURRENT_RESEARCH", Type: model.TypeNumber, Required: true, HasDefault: true, Default: "20", Docs: "Max sessions"},
			{Name: "GEMINI_API_KEY", Type: model.TypeString, Sensitive: true, Docs: "Gemini API key"},
		},
	}

	var buf bytes.Buffer
	if err := GeneratePython(&buf, schema); err != nil {
		t.Fatal(err)
	}

	out := buf.String()

	// Check frozen dataclass.
	if !strings.Contains(out, "@dataclass(frozen=True)") {
		t.Error("missing frozen dataclass")
	}

	// Check pydantic SecretStr import.
	if !strings.Contains(out, "from pydantic import SecretStr") {
		t.Error("missing SecretStr import")
	}

	// Check sensitive field uses SecretStr.
	if !strings.Contains(out, "database_url: SecretStr") {
		t.Errorf("missing database_url SecretStr field:\n%s", out)
	}

	// Check non-sensitive numeric field is plain int.
	if !strings.Contains(out, "port: int") {
		t.Error("missing port field")
	}

	// Check optional sensitive uses SecretStr | None.
	if !strings.Contains(out, "gemini_api_key: SecretStr | None = None") {
		t.Errorf("missing optional gemini_api_key SecretStr field:\n%s", out)
	}

	// Check from_env wraps sensitive values.
	if !strings.Contains(out, "SecretStr(database_url)") {
		t.Errorf("missing SecretStr wrapping in from_env:\n%s", out)
	}

	// Check from_env.
	if !strings.Contains(out, "def from_env(cls) -> Config:") {
		t.Error("missing from_env classmethod")
	}

	// Check default.
	if !strings.Contains(out, `os.environ.get("PORT", "8080")`) {
		t.Error("missing PORT default")
	}

	// Check required validation.
	if !strings.Contains(out, `"DATABASE_URL is required"`) {
		t.Error("missing DATABASE_URL required check")
	}

	// Check int coercion.
	if !strings.Contains(out, "port = int(port)") {
		t.Error("missing port int coercion")
	}
}

func TestGeneratePythonNoSensitive(t *testing.T) {
	schema := &model.SchemaFile{
		Vars: []model.VarDef{
			{Name: "PORT", Type: model.TypePort, Required: true, HasDefault: true, Default: "3000"},
		},
	}

	var buf bytes.Buffer
	if err := GeneratePython(&buf, schema); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	// No sensitive vars means no SecretStr import.
	if strings.Contains(out, "SecretStr") {
		t.Error("should not import SecretStr when no sensitive vars")
	}
}

func TestGeneratePythonOptional(t *testing.T) {
	schema := &model.SchemaFile{
		Vars: []model.VarDef{
			{Name: "ANTHROPIC_API_KEY", Type: model.TypeString, Sensitive: true, Docs: "API key"},
		},
	}

	var buf bytes.Buffer
	if err := GeneratePython(&buf, schema); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, `os.environ.get("ANTHROPIC_API_KEY") or None`) {
		t.Errorf("missing optional get:\n%s", out)
	}
	// Optional sensitive should wrap conditionally.
	if !strings.Contains(out, "SecretStr(anthropic_api_key) if anthropic_api_key is not None else None") {
		t.Errorf("missing conditional SecretStr wrap:\n%s", out)
	}
}

func TestGeneratePythonEnum(t *testing.T) {
	schema := &model.SchemaFile{
		Vars: []model.VarDef{
			{Name: "NODE_ENV", Type: model.TypeEnum, EnumValues: []string{"development", "production", "test"}, Required: true, Docs: "Environment"},
		},
	}

	var buf bytes.Buffer
	if err := GeneratePython(&buf, schema); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, `"development", "production", "test"`) {
		t.Errorf("missing enum validation:\n%s", out)
	}
}
