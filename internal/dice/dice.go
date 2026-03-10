package dice

import (
	"math/rand"
	"strconv"
	"strings"
)

// D rolls dice in the format used by the Perl app.
// Simple int: D(6) = 1d6
// "6x2" = 2d6
// "6p1" = 1d6+1
func D(spec string) int {
	sides := 0
	times := 1
	plus := 0

	if strings.Contains(spec, "x") {
		parts := strings.SplitN(spec, "x", 2)
		sides, _ = strconv.Atoi(parts[0])
		times, _ = strconv.Atoi(parts[1])
	} else if strings.Contains(spec, "p") {
		parts := strings.SplitN(spec, "p", 2)
		sides, _ = strconv.Atoi(parts[0])
		plus, _ = strconv.Atoi(parts[1])
	} else {
		sides, _ = strconv.Atoi(spec)
	}

	total := 0
	for i := 0; i < times; i++ {
		total += rand.Intn(sides) + 1
	}
	total += plus
	return total
}

// DInt rolls a single die with the given number of sides.
func DInt(sides int) int {
	return rand.Intn(sides) + 1
}
