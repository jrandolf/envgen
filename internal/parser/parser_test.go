package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jrandolf/envgen/internal/model"
)

func writeSchema(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, ".env.schema")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestParseHeader(t *testing.T) {
	schema := `# @defaultSensitive=false
# @defaultRequired=true
# ---

# @type=port
# @docs("HTTP port")
PORT=3000
`
	s, err := ParseFile(writeSchema(t, schema))
	if err != nil {
		t.Fatal(err)
	}
	if s.DefaultSensitive != false {
		t.Error("expected defaultSensitive=false")
	}
	if s.DefaultRequired != true {
		t.Error("expected defaultRequired=true")
	}
}

func TestParseRequired(t *testing.T) {
	schema := `# @defaultSensitive=false
# @defaultRequired=true
# ---

# @sensitive @type=url
# @docs("Database URL")
DATABASE_URL=

# @type=port
# @docs("HTTP port")
PORT=3000
`
	s, err := ParseFile(writeSchema(t, schema))
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Vars) != 2 {
		t.Fatalf("expected 2 vars, got %d", len(s.Vars))
	}

	db := s.Vars[0]
	if db.Name != "DATABASE_URL" {
		t.Errorf("expected DATABASE_URL, got %s", db.Name)
	}
	if !db.Required {
		t.Error("DATABASE_URL should be required")
	}
	if !db.Sensitive {
		t.Error("DATABASE_URL should be sensitive")
	}
	if db.Type != model.TypeURL {
		t.Errorf("DATABASE_URL type should be url, got %s", db.Type)
	}
	if db.HasDefault {
		t.Error("DATABASE_URL should not have a default")
	}

	port := s.Vars[1]
	if port.Name != "PORT" {
		t.Errorf("expected PORT, got %s", port.Name)
	}
	if !port.Required {
		t.Error("PORT should be required")
	}
	if port.Sensitive {
		t.Error("PORT should not be sensitive")
	}
	if port.Type != model.TypePort {
		t.Errorf("PORT type should be port, got %s", port.Type)
	}
	if !port.HasDefault || port.Default != "3000" {
		t.Errorf("PORT should have default 3000, got %q (has=%v)", port.Default, port.HasDefault)
	}
}

func TestParseOptional(t *testing.T) {
	schema := `# @defaultSensitive=false
# @defaultRequired=true
# ---

# @optional @sensitive
# @docs("API key")
ANTHROPIC_API_KEY=
`
	s, err := ParseFile(writeSchema(t, schema))
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Vars) != 1 {
		t.Fatalf("expected 1 var, got %d", len(s.Vars))
	}

	v := s.Vars[0]
	if v.Required {
		t.Error("should be optional")
	}
	if !v.Sensitive {
		t.Error("should be sensitive")
	}
}

func TestParseDocs(t *testing.T) {
	schema := `# @defaultSensitive=false
# @defaultRequired=true
# ---

# @type=url
# @docs("NATS server URL", https://docs.nats.io/running-a-nats-service/clients)
NATS_URL=
`
	s, err := ParseFile(writeSchema(t, schema))
	if err != nil {
		t.Fatal(err)
	}

	v := s.Vars[0]
	if v.Docs != "NATS server URL" {
		t.Errorf("unexpected docs: %q", v.Docs)
	}
	if v.DocsURL != "https://docs.nats.io/running-a-nats-service/clients" {
		t.Errorf("unexpected docs URL: %q", v.DocsURL)
	}
}

func TestParseEnum(t *testing.T) {
	schema := `# @defaultSensitive=false
# @defaultRequired=true
# ---

# @type=enum(development, production, test)
# @docs("Node environment")
NODE_ENV=
`
	s, err := ParseFile(writeSchema(t, schema))
	if err != nil {
		t.Fatal(err)
	}

	v := s.Vars[0]
	if v.Type != model.TypeEnum {
		t.Errorf("expected enum type, got %s", v.Type)
	}
	if len(v.EnumValues) != 3 {
		t.Fatalf("expected 3 enum values, got %d", len(v.EnumValues))
	}
	expected := []string{"development", "production", "test"}
	for i, e := range expected {
		if v.EnumValues[i] != e {
			t.Errorf("enum[%d]: expected %q, got %q", i, e, v.EnumValues[i])
		}
	}
}

func TestParseNumber(t *testing.T) {
	schema := `# @defaultSensitive=false
# @defaultRequired=true
# ---

# @type=number
# @docs("Embedding dimension")
EMBEDDING_DIMENSION=3072
`
	s, err := ParseFile(writeSchema(t, schema))
	if err != nil {
		t.Fatal(err)
	}

	v := s.Vars[0]
	if v.Type != model.TypeNumber {
		t.Errorf("expected number type, got %s", v.Type)
	}
	if !v.HasDefault || v.Default != "3072" {
		t.Errorf("expected default 3072, got %q (has=%v)", v.Default, v.HasDefault)
	}
}

func TestParseFullSchema(t *testing.T) {
	// Test with a realistic schema.
	schema := `# @defaultSensitive=false
# @defaultRequired=true
# ---

# @type=url
# @docs("NATS server URL", https://docs.nats.io/running-a-nats-service/clients)
NATS_URL=

# @type=number
# @docs("Number of JetStream stream replicas")
NATS_REPLICAS=

# @optional @sensitive
# @docs("NKey seed file path")
NATS_NKEY_SEED_PATH=

# @sensitive @type=url
# @docs("PostgreSQL connection string")
DATABASE_URL=

# @type=port
# @docs("HTTP port")
PORT=50052
`
	s, err := ParseFile(writeSchema(t, schema))
	if err != nil {
		t.Fatal(err)
	}

	if len(s.Vars) != 5 {
		t.Fatalf("expected 5 vars, got %d", len(s.Vars))
	}

	// NATS_URL: required, url, no default.
	if s.Vars[0].Name != "NATS_URL" || !s.Vars[0].Required || s.Vars[0].HasDefault {
		t.Error("NATS_URL mismatch")
	}

	// NATS_REPLICAS: required, number, no default.
	if s.Vars[1].Name != "NATS_REPLICAS" || s.Vars[1].Type != model.TypeNumber || s.Vars[1].HasDefault {
		t.Error("NATS_REPLICAS mismatch")
	}

	// NATS_NKEY_SEED_PATH: optional, sensitive.
	if s.Vars[2].Name != "NATS_NKEY_SEED_PATH" || s.Vars[2].Required || !s.Vars[2].Sensitive {
		t.Error("NATS_NKEY_SEED_PATH mismatch")
	}

	// DATABASE_URL: required, sensitive, url.
	if s.Vars[3].Name != "DATABASE_URL" || !s.Vars[3].Sensitive || s.Vars[3].Type != model.TypeURL {
		t.Error("DATABASE_URL mismatch")
	}

	// PORT: required, port, default 50052.
	if s.Vars[4].Name != "PORT" || s.Vars[4].Type != model.TypePort || s.Vars[4].Default != "50052" {
		t.Error("PORT mismatch")
	}
}
