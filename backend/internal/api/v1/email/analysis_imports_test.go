package email

import (
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"
)

// TestEmailPackageNeverImportsAIAgent enforces a hard rule of the e-mail threat
// model: the analysis pipeline drives the LLM with text in / JSON out only and
// MUST NOT use the tool-calling agent. If any non-test file in this package ever
// imports pkg/ai/agent, an adversarial e-mail could reach a tool — so we fail
// the build here instead.
func TestEmailPackageNeverImportsAIAgent(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read package dir: %v", err)
	}

	fset := token.NewFileSet()
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}

		file, parseErr := parser.ParseFile(fset, name, nil, parser.ImportsOnly)
		if parseErr != nil {
			t.Fatalf("parse %s: %v", name, parseErr)
		}
		for _, imp := range file.Imports {
			path := strings.Trim(imp.Path.Value, `"`)
			if strings.Contains(path, "pkg/ai/agent") {
				t.Fatalf("%s imports %s: the e-mail pipeline must never use the AI agent", name, path)
			}
		}
	}
}
