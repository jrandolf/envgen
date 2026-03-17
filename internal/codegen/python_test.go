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

	// Check field types (required first, optional last).
	if !strings.Contains(out, "database_url: str") {
		t.Error("missing database_url field")
	}
	if !strings.Contains(out, "port: int") {
		t.Error("missing port field")
	}
	if !strings.Contains(out, "gemini_api_key: str | None = None") {
		t.Error("missing optional gemini_api_key field")
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
