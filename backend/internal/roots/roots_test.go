package roots

import (
	"path/filepath"
	"testing"

	"nas-go/api/internal/config"
)

func setupRoots(t *testing.T, entryPoint string, list []Root) {
	t.Helper()
	originalEntryPoint := config.AppConfig.EntryPoint
	t.Cleanup(func() {
		config.AppConfig.EntryPoint = originalEntryPoint
		Reset()
	})
	config.AppConfig.EntryPoint = entryPoint
	Set(list)
}

func abs(parts ...string) string {
	return filepath.Join(append([]string{string(filepath.Separator)}, parts...)...)
}

func TestFallbackToEntryPointWhenRegistryEmpty(t *testing.T) {
	setupRoots(t, abs("data"), nil)

	enabled := Enabled()
	if len(enabled) != 1 || enabled[0].Path != abs("data") {
		t.Fatalf("expected the ENTRY_POINT fallback root, got %+v", enabled)
	}

	primary, ok := Primary()
	if !ok || primary.Path != abs("data") {
		t.Fatalf("expected primary fallback, got %+v ok=%v", primary, ok)
	}

	// Legacy behavior preserved byte-for-byte.
	if got := ToRelativePath(abs("data", "fotos")); got != "/fotos" {
		t.Fatalf("ToRelativePath legacy = %q", got)
	}
	if got := ToRelativePath(abs("data")); got != "/" {
		t.Fatalf("ToRelativePath root = %q", got)
	}
	if got := ToAbsolutePath("/fotos"); got != abs("data", "fotos") {
		t.Fatalf("ToAbsolutePath legacy = %q", got)
	}
	if got := ToAbsolutePath("/"); got != abs("data") {
		t.Fatalf("ToAbsolutePath root = %q", got)
	}
}

func TestEnabledFiltersDisabledRoots(t *testing.T) {
	setupRoots(t, abs("data"), []Root{
		{ID: 1, Path: abs("data"), Label: "data", Enabled: true},
		{ID: 2, Path: abs("midia"), Label: "Midia", Enabled: false},
	})

	enabled := Enabled()
	if len(enabled) != 1 || enabled[0].ID != 1 {
		t.Fatalf("expected only the enabled root, got %+v", enabled)
	}
}

func TestOwnerOfPrefersLongestMatchAndRejectsSiblingPrefix(t *testing.T) {
	setupRoots(t, abs("data"), []Root{
		{ID: 1, Path: abs("data"), Label: "data", Enabled: true},
		{ID: 2, Path: abs("midia"), Label: "Midia", Enabled: true},
	})

	owner, ok := OwnerOf(abs("midia", "filmes", "x.mp4"))
	if !ok || owner.ID != 2 {
		t.Fatalf("expected root 2, got %+v ok=%v", owner, ok)
	}

	// A sibling sharing the string prefix is NOT inside the root.
	if _, ok := OwnerOf(abs("midia-extra", "x.mp4")); ok {
		t.Fatalf("sibling dir sharing the prefix must not match")
	}

	if _, ok := OwnerOf(abs("fora", "x.txt")); ok {
		t.Fatalf("path outside every root must not match")
	}
}

func TestRelativePathsAcrossRoots(t *testing.T) {
	setupRoots(t, abs("data"), []Root{
		{ID: 1, Path: abs("data"), Label: "data", Enabled: true},
		{ID: 2, Path: abs("midia"), Label: "Midia", Enabled: true},
	})

	// Primary root: legacy shape.
	if got := ToRelativePath(abs("data", "docs", "a.txt")); got != "/docs/a.txt" {
		t.Fatalf("primary relative = %q", got)
	}

	// Additional root: label-prefixed shape.
	if got := ToRelativePath(abs("midia", "filmes", "x.mp4")); got != "/Midia/filmes/x.mp4" {
		t.Fatalf("additional relative = %q", got)
	}
	if got := ToRelativePath(abs("midia")); got != "/Midia" {
		t.Fatalf("additional root itself = %q", got)
	}
}

func TestAbsolutePathsAcrossRoots(t *testing.T) {
	setupRoots(t, abs("data"), []Root{
		{ID: 1, Path: abs("data"), Label: "data", Enabled: true},
		{ID: 2, Path: abs("midia"), Label: "Midia", Enabled: true},
	})

	if got := ToAbsolutePath("/Midia/filmes/x.mp4"); got != abs("midia", "filmes", "x.mp4") {
		t.Fatalf("additional absolute = %q", got)
	}
	if got := ToAbsolutePath("/Midia"); got != abs("midia") {
		t.Fatalf("additional root absolute = %q", got)
	}
	// Anything else falls under the primary root, like the old join.
	if got := ToAbsolutePath("/docs/a.txt"); got != abs("data", "docs", "a.txt") {
		t.Fatalf("primary absolute = %q", got)
	}
}

func TestResolveAbsoluteEnforcesContainment(t *testing.T) {
	setupRoots(t, abs("data"), []Root{
		{ID: 1, Path: abs("data"), Label: "data", Enabled: true},
		{ID: 2, Path: abs("midia"), Label: "Midia", Enabled: true},
	})

	resolved, err := ResolveAbsolute(abs("midia", "filmes"))
	if err != nil || resolved != abs("midia", "filmes") {
		t.Fatalf("absolute inside root: %q err=%v", resolved, err)
	}

	resolved, err = ResolveAbsolute("docs/a.txt")
	if err != nil || resolved != abs("data", "docs", "a.txt") {
		t.Fatalf("relative under primary: %q err=%v", resolved, err)
	}

	resolved, err = ResolveAbsolute("")
	if err != nil || resolved != abs("data") {
		t.Fatalf("empty resolves to primary: %q err=%v", resolved, err)
	}

	if _, err := ResolveAbsolute(abs("fora", "x.txt")); err == nil {
		t.Fatalf("absolute outside every root must fail")
	}

	// Path-traversal out of the roots must fail too.
	if _, err := ResolveAbsolute("../../etc/passwd"); err == nil {
		t.Fatalf("traversal outside every root must fail")
	}
}

func TestResolveAbsoluteWithoutAnyRoot(t *testing.T) {
	setupRoots(t, "", nil)

	if _, err := ResolveAbsolute("/qualquer"); err == nil {
		t.Fatalf("expected error without any configured root")
	}
	if len(Enabled()) != 0 {
		t.Fatalf("expected no enabled roots")
	}
	if _, ok := Primary(); ok {
		t.Fatalf("expected no primary root")
	}
}
