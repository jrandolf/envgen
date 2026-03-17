package model

// VarType represents the type of a schema variable.
type VarType int

const (
	TypeString VarType = iota
	TypePort
	TypeNumber
	TypeURL
	TypeEmail
	TypeEnum
)

func (t VarType) String() string {
	switch t {
	case TypePort:
		return "port"
	case TypeNumber:
		return "number"
	case TypeURL:
		return "url"
	case TypeEmail:
		return "email"
	case TypeEnum:
		return "enum"
	default:
		return "string"
	}
}

// VarDef represents a single environment variable definition.
type VarDef struct {
	Name       string
	Type       VarType
	EnumValues []string // for enum(a, b, c)
	Required   bool     // inverse of @optional
	Sensitive  bool     // @sensitive
	Default    string   // value after =
	HasDefault bool     // true if any value after = (including empty string)
	Docs       string
	DocsURL    string
}

// SchemaFile represents a parsed .env.schema file.
type SchemaFile struct {
	DefaultSensitive bool
	DefaultRequired  bool
	Vars             []VarDef
}
