package conflux

import (
	"strings"
	"testing"
)

const sampleData = `Alpha Beta Gamma Delta Epsilon
Zeta Eta Theta Iota Kappa Lambda`

func TestGenerate_ReturnsRequestedCount(t *testing.T) {
	names := Generate(sampleData, 10)
	if len(names) != 10 {
		t.Errorf("Generate returned %d names; want 10", len(names))
	}
}

func TestGenerate_NamesHaveValidLength(t *testing.T) {
	names := Generate(sampleData, 50)
	for i, name := range names {
		if len(name) > maxLength {
			t.Errorf("Name %d (%q) exceeds max length %d", i, name, maxLength)
		}
		if len(name) == 0 {
			t.Errorf("Name %d is empty", i)
		}
	}
}

func TestGenerate_EmptyInput(t *testing.T) {
	names := Generate("", 10)
	if names != nil {
		t.Errorf("Generate with empty input returned %v; want nil", names)
	}
}

func TestGenerate_CommentsStripped(t *testing.T) {
	data := "Hello World # this is a comment\nFoo Bar"
	names := Generate(data, 5)
	if len(names) != 5 {
		t.Errorf("Generate returned %d names; want 5", len(names))
	}
}

func TestGenerate_ProducesNames(t *testing.T) {
	names := Generate(sampleData, 10)
	if len(names) != 10 {
		t.Fatalf("Generate returned %d names; want 10", len(names))
	}
	for i, name := range names {
		trimmed := strings.TrimSpace(name)
		if len(trimmed) == 0 {
			t.Errorf("Name %d is empty after trimming", i)
		}
	}
}
