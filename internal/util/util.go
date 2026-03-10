package util

import (
	"fmt"
	"math/rand"
	"strings"
)

const SeedMax = 4294967296 // 2^32

func CellID(col, row int) string {
	return fmt.Sprintf("0%d0%d", col, row)
}

func SysName(star, cell string) string {
	return "System:" + strings.ToUpper(star)
}

func RandomSeed() int64 {
	seed := int64(rand.Intn(int(SeedMax)))
	if seed == 0 {
		seed = 1
	}
	return seed
}

func Tagify(txt string) string {
	words := strings.Fields(txt)
	var result strings.Builder
	for _, w := range words {
		if len(w) > 0 {
			result.WriteString(strings.ToUpper(w[:1]) + strings.ToLower(w[1:]))
		}
	}
	return result.String()
}

func Filename(txt string) string {
	return strings.ReplaceAll(txt, " ", "_")
}

func TokenizeSeed(seed int64) string {
	if seed == 0 {
		return "0"
	}
	base := int64(36)
	var pieces []byte
	s := seed
	for s > 0 {
		num := s % base
		s = s / base
		var ch byte
		if num > 9 {
			ch = byte('A') + byte(num-10)
		} else {
			ch = byte('0') + byte(num)
		}
		pieces = append([]byte{ch}, pieces...)
	}
	return string(pieces)
}

func UntokenizeSeed(token string) int64 {
	chars := []byte(strings.ToUpper(token))
	seed := int64(0)
	mult := int64(1)
	for i := len(chars) - 1; i >= 0; i-- {
		ch := chars[i]
		var num int64
		if ch >= '0' && ch <= '9' {
			num = int64(ch - '0')
		} else {
			num = int64(ch-'A') + 10
		}
		seed += num * mult
		mult *= 36
	}
	if seed >= SeedMax {
		seed = SeedMax - 1
	}
	if seed < 1 {
		seed = 1
	}
	return seed
}
