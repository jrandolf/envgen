package parser

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/jrandolf/envgen/internal/model"
)

var (
	headerDirectiveRe = regexp.MustCompile(`^#\s*@(\w+)=(\w+)`)
	// Matches @name, @name=value, @name(args), or @name=enum(a, b, c).
	annotationRe = regexp.MustCompile(`@(\w+)(?:=(\w+(?:\([^)]*\))?))?(?:\(([^)]*)\))?`)
	typeEnumRe = regexp.MustCompile(`^enum\((.+)\)$`)
	varLineRe  = regexp.MustCompile(`^([A-Z][A-Z0-9_]*)=(.*)$`)
)

// ParseFile parses a .env.schema file and returns the schema definition.
func ParseFile(path string) (*model.SchemaFile, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open schema: %w", err)
	}
	defer f.Close()

	schema := &model.SchemaFile{
		DefaultSensitive: false,
		DefaultRequired:  true,
	}

	scanner := bufio.NewScanner(f)
	inHeader := true
	var commentBlock []string

	for scanner.Scan() {
		line := scanner.Text()

		// Header section: parse directives until we hit "# ---"
		if inHeader {
			if strings.TrimSpace(line) == "# ---" {
				inHeader = false
				continue
			}
			if m := headerDirectiveRe.FindStringSubmatch(line); m != nil {
				switch m[1] {
				case "defaultSensitive":
					schema.DefaultSensitive = m[2] == "true"
				case "defaultRequired":
					schema.DefaultRequired = m[2] == "true"
				}
			}
			continue
		}

		// Blank line resets the comment block.
		if strings.TrimSpace(line) == "" {
			commentBlock = nil
			continue
		}

		// Comment line: accumulate.
		if strings.HasPrefix(line, "#") {
			commentBlock = append(commentBlock, line)
			continue
		}

		// Variable line: KEY=value
		if m := varLineRe.FindStringSubmatch(line); m != nil {
			v := parseVar(m[1], m[2], commentBlock, schema.DefaultRequired, schema.DefaultSensitive)
			schema.Vars = append(schema.Vars, v)
			commentBlock = nil
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan schema: %w", err)
	}

	return schema, nil
}

func parseVar(name, defaultVal string, comments []string, defaultRequired, defaultSensitive bool) model.VarDef {
	v := model.VarDef{
		Name:      name,
		Type:      model.TypeString,
		Required:  defaultRequired,
		Sensitive: defaultSensitive,
	}

	if defaultVal != "" {
		v.Default = defaultVal
		v.HasDefault = true
	}

	// Parse annotations from comment block.
	block := strings.Join(comments, "\n")

	for _, m := range annotationRe.FindAllStringSubmatch(block, -1) {
		annotation := m[1]
		// m[2] is the =value form, m[3] is the (args) form.
		arg := m[2]
		if arg == "" {
			arg = m[3]
		}

		switch annotation {
		case "optional":
			v.Required = false
		case "sensitive":
			v.Sensitive = true
		case "type":
			v.Type = parseType(arg)
			if v.Type == model.TypeEnum {
				v.EnumValues = parseEnumValues(arg)
			}
		case "docs":
			// Handled separately below.
		case "defaultSensitive", "defaultRequired", "generateTypes":
			// Header directives, skip.
		}
	}

	// Parse @docs separately — manual extraction to handle nested parens in descriptions.
	if docsArg := extractDocsArg(block); docsArg != "" {
		v.Docs, v.DocsURL = parseDocs(docsArg)
	}

	return v
}

func parseType(arg string) model.VarType {
	if strings.HasPrefix(arg, "enum(") || strings.HasPrefix(arg, "enum ") {
		return model.TypeEnum
	}
	switch strings.TrimSpace(arg) {
	case "port":
		return model.TypePort
	case "number":
		return model.TypeNumber
	case "url":
		return model.TypeURL
	case "email":
		return model.TypeEmail
	default:
		return model.TypeString
	}
}

func parseEnumValues(arg string) []string {
	if m := typeEnumRe.FindStringSubmatch(arg); m != nil {
		parts := strings.Split(m[1], ",")
		values := make([]string, 0, len(parts))
		for _, p := range parts {
			values = append(values, strings.TrimSpace(p))
		}
		return values
	}
	return nil
}

// extractDocsArg finds @docs(...) in the block and returns the balanced-paren content.
func extractDocsArg(block string) string {
	idx := strings.Index(block, "@docs(")
	if idx < 0 {
		return ""
	}
	start := idx + len("@docs(")
	depth := 1
	for i := start; i < len(block); i++ {
		switch block[i] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return block[start:i]
			}
		}
	}
	return ""
}

func parseDocs(raw string) (docs, url string) {
	// @docs("description", https://...)
	// or @docs("description")
	raw = strings.TrimSpace(raw)

	// Find the quoted description.
	if len(raw) > 0 && raw[0] == '"' {
		end := strings.Index(raw[1:], "\"")
		if end >= 0 {
			docs = raw[1 : end+1]
			rest := strings.TrimSpace(raw[end+2:])
			// Check for URL after comma.
			if strings.HasPrefix(rest, ",") {
				url = strings.TrimSpace(rest[1:])
			}
			return docs, url
		}
	}

	return raw, ""
}
