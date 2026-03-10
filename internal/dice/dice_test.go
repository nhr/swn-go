package dice

import (
	"math/rand"
	"testing"
)

func TestDInt(t *testing.T) {
	rand.Seed(42)
	for i := 0; i < 1000; i++ {
		result := DInt(6)
		if result < 1 || result > 6 {
			t.Errorf("DInt(6) = %d; want 1-6", result)
		}
	}
}

func TestDInt_SingleSided(t *testing.T) {
	for i := 0; i < 100; i++ {
		result := DInt(1)
		if result != 1 {
			t.Errorf("DInt(1) = %d; want 1", result)
		}
	}
}

func TestD_Simple(t *testing.T) {
	rand.Seed(42)
	for i := 0; i < 1000; i++ {
		result := D("6")
		if result < 1 || result > 6 {
			t.Errorf("D(\"6\") = %d; want 1-6", result)
		}
	}
}

func TestD_Multiply(t *testing.T) {
	rand.Seed(42)
	for i := 0; i < 1000; i++ {
		result := D("6x2")
		if result < 2 || result > 12 {
			t.Errorf("D(\"6x2\") = %d; want 2-12", result)
		}
	}
}

func TestD_Plus(t *testing.T) {
	rand.Seed(42)
	for i := 0; i < 1000; i++ {
		result := D("6p1")
		if result < 2 || result > 7 {
			t.Errorf("D(\"6p1\") = %d; want 2-7", result)
		}
	}
}

func TestD_Distribution(t *testing.T) {
	rand.Seed(42)
	counts := make(map[int]int)
	n := 60000
	for i := 0; i < n; i++ {
		counts[DInt(6)]++
	}
	// Each face should appear roughly 1/6 of the time
	expected := n / 6
	tolerance := expected / 5 // 20% tolerance
	for face := 1; face <= 6; face++ {
		if counts[face] < expected-tolerance || counts[face] > expected+tolerance {
			t.Errorf("DInt(6) face %d appeared %d times; expected ~%d (±%d)", face, counts[face], expected, tolerance)
		}
	}
}

func TestD_2d6_Distribution(t *testing.T) {
	rand.Seed(42)
	counts := make(map[int]int)
	n := 100000
	for i := 0; i < n; i++ {
		counts[D("6x2")]++
	}
	// 7 should be the most common result for 2d6
	maxVal := 0
	maxCount := 0
	for v, c := range counts {
		if c > maxCount {
			maxCount = c
			maxVal = v
		}
	}
	if maxVal != 7 {
		t.Errorf("Most common 2d6 result was %d; expected 7", maxVal)
	}
}
