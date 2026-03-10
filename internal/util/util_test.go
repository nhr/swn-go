package util

import (
	"testing"
)

func TestCellID(t *testing.T) {
	tests := []struct {
		col, row int
		want     string
	}{
		{0, 0, "0000"},
		{3, 5, "0305"},
		{7, 9, "0709"},
	}
	for _, tt := range tests {
		got := CellID(tt.col, tt.row)
		if got != tt.want {
			t.Errorf("CellID(%d, %d) = %q; want %q", tt.col, tt.row, got, tt.want)
		}
	}
}

func TestSysName(t *testing.T) {
	got := SysName("Alpha", "0102")
	want := "System:ALPHA"
	if got != want {
		t.Errorf("SysName(\"Alpha\", \"0102\") = %q; want %q", got, want)
	}
}

func TestTagify(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello world", "HelloWorld"},
		{"UPPER CASE", "UpperCase"},
		{"single", "Single"},
		{"three word test", "ThreeWordTest"},
		{"", ""},
	}
	for _, tt := range tests {
		got := Tagify(tt.input)
		if got != tt.want {
			t.Errorf("Tagify(%q) = %q; want %q", tt.input, got, tt.want)
		}
	}
}

func TestFilename(t *testing.T) {
	got := Filename("Hello World")
	want := "Hello_World"
	if got != want {
		t.Errorf("Filename(\"Hello World\") = %q; want %q", got, want)
	}
}

func TestTokenizeSeed(t *testing.T) {
	tests := []struct {
		seed int64
		want string
	}{
		{0, "0"},
		{1, "1"},
		{10, "A"},
		{35, "Z"},
		{36, "10"},
		{1296, "100"},
	}
	for _, tt := range tests {
		got := TokenizeSeed(tt.seed)
		if got != tt.want {
			t.Errorf("TokenizeSeed(%d) = %q; want %q", tt.seed, got, tt.want)
		}
	}
}

func TestUntokenizeSeed(t *testing.T) {
	tests := []struct {
		token string
		want  int64
	}{
		{"1", 1},
		{"A", 10},
		{"Z", 35},
		{"10", 36},
		{"100", 1296},
	}
	for _, tt := range tests {
		got := UntokenizeSeed(tt.token)
		if got != tt.want {
			t.Errorf("UntokenizeSeed(%q) = %d; want %d", tt.token, got, tt.want)
		}
	}
}

func TestTokenizeUntokenizeRoundTrip(t *testing.T) {
	seeds := []int64{1, 42, 100, 1000, 123456, 4294967295}
	for _, seed := range seeds {
		token := TokenizeSeed(seed)
		got := UntokenizeSeed(token)
		if got != seed {
			t.Errorf("Round trip failed: seed=%d -> token=%q -> %d", seed, token, got)
		}
	}
}

func TestUntokenizeSeed_CaseInsensitive(t *testing.T) {
	upper := UntokenizeSeed("ABC")
	lower := UntokenizeSeed("abc")
	if upper != lower {
		t.Errorf("UntokenizeSeed is not case-insensitive: ABC=%d, abc=%d", upper, lower)
	}
}

func TestUntokenizeSeed_Clamp(t *testing.T) {
	// Overflow: should clamp to SeedMax - 1
	got := UntokenizeSeed("ZZZZZZZZZZ")
	if got != SeedMax-1 {
		t.Errorf("UntokenizeSeed overflow = %d; want %d", got, SeedMax-1)
	}

	// Underflow: empty string parses to 0, should clamp to 1
	got = UntokenizeSeed("")
	if got != 1 {
		t.Errorf("UntokenizeSeed(\"\") = %d; want 1", got)
	}
}

func TestRandomSeed(t *testing.T) {
	for i := 0; i < 100; i++ {
		seed := RandomSeed()
		if seed < 1 || seed >= SeedMax {
			t.Errorf("RandomSeed() = %d; want 1..%d", seed, SeedMax-1)
		}
	}
}
