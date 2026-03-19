# Skip Validation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `SKIP_ENV_VALIDATION` support to all three code generators so that generated config loaders skip the error gate when that env var is set to a non-empty value.

**Architecture:** Each codegen file emits the same validation error collection as before, but the final throw/panic/raise is guarded by a `SKIP_ENV_VALIDATION` check. Missing required vars still get assigned empty string / zero — the gate just won't abort the process. One-line change per language, no structural changes.

**Tech Stack:** Go 1.26, no external dependencies.

---

### Task 1: Add SKIP_ENV_VALIDATION to Go codegen

**Files:**
- Modify: `internal/codegen/golang.go:162-164`
- Test: `internal/codegen/golang_test.go`

- [ ] **Step 1: Write the failing test**

Add to `golang_test.go`:

```go
func TestGenerateGoSkipValidation(t *testing.T) {
	schema := &model.SchemaFile{
		Vars: []model.VarDef{
			{Name: "REQUIRED_VAR", Type: model.TypeString, Required: true},
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
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd /Users/jrandolf/Sources/envgen && go test ./internal/codegen/ -run TestGenerateGoSkipValidation -v
```

Expected: FAIL — `expected SKIP_ENV_VALIDATION check in error gate`

- [ ] **Step 3: Implement the change**

In `internal/codegen/golang.go`, find `writeLoadFunc`. Change the error gate (currently line 162):

```go
// Before:
fmt.Fprintf(w, "\tif len(errs) > 0 {\n")

// After:
fmt.Fprintf(w, "\tif len(errs) > 0 && os.Getenv(\"SKIP_ENV_VALIDATION\") == \"\" {\n")
```

- [ ] **Step 4: Run test to verify it passes**

```bash
cd /Users/jrandolf/Sources/envgen && go test ./internal/codegen/ -run TestGenerateGoSkipValidation -v
```

Expected: PASS

- [ ] **Step 5: Run all Go codegen tests to check for regressions**

```bash
cd /Users/jrandolf/Sources/envgen && go test ./internal/codegen/ -run TestGenerateGo -v
```

Expected: all PASS

- [ ] **Step 6: Commit**

```bash
cd /Users/jrandolf/Sources/envgen && git add internal/codegen/golang.go internal/codegen/golang_test.go
git commit -m "Add SKIP_ENV_VALIDATION support to Go codegen"
```

---

### Task 2: Add SKIP_ENV_VALIDATION to TypeScript codegen

**Files:**
- Modify: `internal/codegen/typescript.go:127`
- Test: `internal/codegen/typescript_test.go`

- [ ] **Step 1: Write the failing test**

Add to `typescript_test.go`:

```go
func TestGenerateTypeScriptSkipValidation(t *testing.T) {
	schema := &model.SchemaFile{
		Vars: []model.VarDef{
			{Name: "REQUIRED_VAR", Type: model.TypeString, Required: true},
		},
	}

	var buf bytes.Buffer
	if err := GenerateTypeScript(&buf, schema); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, `!process.env["SKIP_ENV_VALIDATION"]`) {
		t.Errorf("expected SKIP_ENV_VALIDATION check in error gate:\n%s", out)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd /Users/jrandolf/Sources/envgen && go test ./internal/codegen/ -run TestGenerateTypeScriptSkipValidation -v
```

Expected: FAIL — `expected SKIP_ENV_VALIDATION check in error gate`

- [ ] **Step 3: Implement the change**

In `internal/codegen/typescript.go`, find `writeLoadEnvFunc`. Change the error gate (currently line 127):

```go
// Before:
fmt.Fprintf(w, "  if (errors.length > 0) {\n")

// After:
fmt.Fprintf(w, "  if (errors.length > 0 && !process.env[\"SKIP_ENV_VALIDATION\"]) {\n")
```

- [ ] **Step 4: Run test to verify it passes**

```bash
cd /Users/jrandolf/Sources/envgen && go test ./internal/codegen/ -run TestGenerateTypeScriptSkipValidation -v
```

Expected: PASS

- [ ] **Step 5: Run all TypeScript codegen tests to check for regressions**

```bash
cd /Users/jrandolf/Sources/envgen && go test ./internal/codegen/ -run TestGenerateTypeScript -v
```

Expected: all PASS

- [ ] **Step 6: Commit**

```bash
cd /Users/jrandolf/Sources/envgen && git add internal/codegen/typescript.go internal/codegen/typescript_test.go
git commit -m "Add SKIP_ENV_VALIDATION support to TypeScript codegen"
```

---

### Task 3: Add SKIP_ENV_VALIDATION to Python codegen

**Files:**
- Modify: `internal/codegen/python.go:109`
- Test: `internal/codegen/python_test.go`

- [ ] **Step 1: Write the failing test**

Add to `python_test.go`:

```go
func TestGeneratePythonSkipValidation(t *testing.T) {
	schema := &model.SchemaFile{
		Vars: []model.VarDef{
			{Name: "REQUIRED_VAR", Type: model.TypeString, Required: true},
		},
	}

	var buf bytes.Buffer
	if err := GeneratePython(&buf, schema); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, `not os.environ.get("SKIP_ENV_VALIDATION")`) {
		t.Errorf("expected SKIP_ENV_VALIDATION check in error gate:\n%s", out)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd /Users/jrandolf/Sources/envgen && go test ./internal/codegen/ -run TestGeneratePythonSkipValidation -v
```

Expected: FAIL — `expected SKIP_ENV_VALIDATION check in error gate`

- [ ] **Step 3: Implement the change**

In `internal/codegen/python.go`, find the error gate (currently line 109):

```go
// Before:
fmt.Fprintf(w, "        if errors:\n")

// After:
fmt.Fprintf(w, "        if errors and not os.environ.get(\"SKIP_ENV_VALIDATION\"):\n")
```

- [ ] **Step 4: Run test to verify it passes**

```bash
cd /Users/jrandolf/Sources/envgen && go test ./internal/codegen/ -run TestGeneratePythonSkipValidation -v
```

Expected: PASS

- [ ] **Step 5: Run all Python codegen tests to check for regressions**

```bash
cd /Users/jrandolf/Sources/envgen && go test ./internal/codegen/ -run TestGeneratePython -v
```

Expected: all PASS

- [ ] **Step 6: Commit**

```bash
cd /Users/jrandolf/Sources/envgen && git add internal/codegen/python.go internal/codegen/python_test.go
git commit -m "Add SKIP_ENV_VALIDATION support to Python codegen"
```

---

### Task 4: Full test suite + README

**Files:**
- Read: `README.md`
- Modify: `README.md`

- [ ] **Step 1: Run the full test suite**

```bash
cd /Users/jrandolf/Sources/envgen && go test ./...
```

Expected: all PASS

- [ ] **Step 2: Add SKIP_ENV_VALIDATION to README (only if Step 1 passed)**

Read `README.md`, then add a section documenting the env var. Place it near any existing section on validation or runtime behavior. Use this content (paste as plain markdown, not wrapped in a code block):

    ## Skipping validation

    Set `SKIP_ENV_VALIDATION` to any non-empty value to suppress validation errors at startup.
    This is useful in test environments where not all variables are populated:

        SKIP_ENV_VALIDATION=true node my-app.js

    When set, missing required variables receive empty string / zero values and the process
    continues. Type and enum checks are still collected but not raised.

- [ ] **Step 3: Commit**

```bash
cd /Users/jrandolf/Sources/envgen && git add README.md
git commit -m "Document SKIP_ENV_VALIDATION in README"
```
