package conflux

import (
	"math/rand"
	"strings"
)

const (
	minLength = 3
	maxLength = 7
)

// Generate produces alien names using the language confluxer algorithm.
// data is the raw text content of a language data file.
func Generate(data string, number int) []string {
	// Parse input data
	var chars []byte
	lines := strings.Split(data, "\n")
	for _, line := range lines {
		// Strip comments
		if idx := strings.Index(line, "#"); idx >= 0 {
			line = line[:idx]
		}
		if len(line) == 0 {
			continue
		}
		line = strings.TrimSpace(line)
		line = strings.Join(strings.Fields(line), " ")
		chars = append(chars, ' ')
		chars = append(chars, []byte(line)...)
	}

	// Loop around
	if len(chars) > 1 {
		chars = append(chars, ' ', chars[1])
	}

	// Build pair hash
	hash := make(map[string]string)
	var startPairs []string

	for len(chars) > 2 {
		first := chars[0]
		second := chars[1]
		third := chars[2]
		pair := string([]byte{first, second})
		hash[pair] += string(third)
		if first == ' ' {
			startPairs = append(startPairs, string([]byte{second, third}))
		}
		chars = chars[1:]
	}

	if len(startPairs) == 0 {
		return nil
	}

	// Generate words
	words := make([]string, 0, number)
	current := startPairs[rand.Intn(len(startPairs))]
	for i := 0; i < number; i++ {
		current = newWord(current[len(current)-2:], hash)
		if len(current) > maxLength {
			current = current[:maxLength]
		}
		words = append(words, current)
	}

	return words
}

func newWord(seed string, hash map[string]string) string {
	word := seed
	for {
		pair := word[len(word)-2:]
		candidates, ok := hash[pair]
		if !ok || len(candidates) == 0 {
			if len(word) > minLength {
				return word
			}
			// restart with last char + random
			if len(hash) > 0 {
				for k, v := range hash {
					if len(v) > 0 {
						word = string(word[len(word)-1]) + string(v[rand.Intn(len(v))])
						break
					}
					_ = k
				}
			}
			continue
		}
		letter := candidates[rand.Intn(len(candidates))]
		if word[len(word)-1] == ' ' {
			if len(word) > minLength {
				return word
			}
			word = string(word[len(word)-1]) + string(letter)
		} else {
			word = strings.TrimLeft(word, " ")
			word += string(letter)
		}
	}
}
