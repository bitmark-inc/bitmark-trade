package account

import (
	"errors"
	"fmt"

	"github.com/bitmark-inc/bitmark-sdk-go/account/bip39"
	"golang.org/x/text/language"
)

// 0..10 bit masks
var masks = []int{0, 1, 3, 7, 15, 31, 63, 127, 255, 511, 1023}

// convert a binary of 33 bytes to a phrase of 24 worhs
func bytesToTwentyFourWords(input []byte) ([]string, error) {
	if 33 != len(input) {
		return nil, fmt.Errorf("input length: %d expected: 33", len(input))
	}

	phrase := make([]string, 0, 24)
	accumulator := 0
	bits := 0
	n := 0
	for i := 0; i < len(input); i++ {
		accumulator = accumulator<<8 + int(input[i])
		bits += 8
		if bits >= 11 {
			bits -= 11 // [ 11 bits] [offset bits]

			n++
			index := accumulator >> uint(bits)
			accumulator &= masks[bits]
			word := bip39.English[index]
			phrase = append(phrase, word)
		}
	}
	if 24 != len(phrase) {
		return nil, fmt.Errorf("only %d words expected 24", len(phrase))
	}
	return phrase, nil
}

// 24 words to 33 bytes
func twentyFourWordsToBytes(words []string) ([]byte, error) {
	if 24 != len(words) {
		return nil, fmt.Errorf("number of words: %d expected: 24", len(words))
	}

	databytes := make([]byte, 0, 33)

	remainder := 0
	bits := 0
	for _, word := range words {
		n := -1
	loop:
		for i, bip := range bip39.English {
			if word == bip {
				n = i
				break loop
			}
		}
		if n < 0 {
			return nil, fmt.Errorf("invalid word: %q", word)
		}
		remainder = remainder<<11 + n
		for bits += 11; bits >= 8; bits -= 8 {
			a := 0xff & (remainder >> uint(bits-8))
			databytes = append(databytes, byte(a))
		}
		remainder &= masks[bits]
	}
	if 33 != len(databytes) {
		return nil, fmt.Errorf("only converted: %d bytes expected: 33", len(databytes))
	}
	return databytes, nil
}

// 17.5 bytes to 12 words
func bytesToTwelveWords(input []byte, dict []string) ([]string, error) {
	phrase := make([]string, 0, 12)
	accumulator := 0
	bits := 0
	n := 0
	for i := 0; i < len(input); i += 1 {
		accumulator = accumulator<<8 + int(input[i])
		bits += 8
		if bits >= 11 {
			bits -= 11 // [ 11 bits] [offset bits]

			n += 1
			index := accumulator >> uint(bits)
			accumulator &= masks[bits]

			phrase = append(phrase, dict[index])
		}
	}

	if 12 != len(phrase) {
		return nil, fmt.Errorf("oly %d words expected 12", len(phrase))
	}

	return phrase, nil
}

// 12 words to 17.5 bytes
func twelveWordsToByteswords(words []string, dict []string) ([]byte, error) {
	seed := make([]byte, 0, 17)

	remainder := 0
	bits := 0
	for _, word := range words {
		n := -1
	loop:
		for i, bip := range dict {
			if word == bip {
				n = i
				break loop
			}
		}
		if n < 0 {
			return nil, fmt.Errorf("invalid word: %q", word)
		}
		remainder = remainder<<11 + n
		for bits += 11; bits >= 8; bits -= 8 {
			a := 0xff & (remainder >> uint(bits-8))
			seed = append(seed, byte(a))
		}
		remainder &= masks[bits]
	}

	// check that the whole 16 bytes are converted and the final nibble remains to be packed
	if 4 != bits || 16 != len(seed) {
		return nil, fmt.Errorf("only converted: %d bytes expected: 16.5", len(seed))
	}

	// justify final 4 bits to high nibble, low nibble is zero
	seed = append(seed, byte(remainder<<4))
	return seed, nil
}

func getBIP39Dict(lang language.Tag) ([]string, error) {
	switch lang {
	case language.AmericanEnglish:
		return bip39.English, nil
	case language.TraditionalChinese:
		return bip39.TraditionalChinese, nil
	default:
		return nil, errors.New("language not supported")
	}
}
