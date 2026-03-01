package highlight

import (
	"reflect"
	"testing"
	"unicode/utf8"
)

func TestIsMark(t *testing.T) {
	tests := []struct {
		name string
		r    rune
		want bool
	}{
		{
			name: "ASCII letter",
			r:    'a',
			want: false,
		},
		{
			name: "ASCII number",
			r:    '5',
			want: false,
		},
		{
			name: "combining acute accent",
			r:    '\u0301',
			want: true,
		},
		{
			name: "combining grave accent",
			r:    '\u0300',
			want: true,
		},
		{
			name: "combining tilde",
			r:    '\u0303',
			want: true,
		},
		{
			name: "combining diaeresis",
			r:    '\u0308',
			want: true,
		},
		{
			name: "space",
			r:    ' ',
			want: false,
		},
		{
			name: "newline",
			r:    '\n',
			want: false,
		},
		{
			name: "CJK character",
			r:    '世',
			want: false,
		},
		{
			name: "emoji",
			r:    '😀',
			want: false,
		},
		{
			name: "null rune",
			r:    '\x00',
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isMark(tt.r); got != tt.want {
				t.Errorf("isMark(%q) = %v, want %v", tt.r, got, tt.want)
			}
		})
	}
}

func TestDecodeCharacter(t *testing.T) {
	tests := []struct {
		name      string
		input     []byte
		wantRune  rune
		wantCombc []rune
		wantSize  int
	}{
		{
			name:      "simple ASCII",
			input:     []byte("hello"),
			wantRune:  'h',
			wantCombc: nil,
			wantSize:  1,
		},
		{
			name:      "simple unicode",
			input:     []byte("世界"),
			wantRune:  '世',
			wantCombc: nil,
			wantSize:  3,
		},
		{
			name:      "character with combining accent",
			input:     []byte("e\u0301"), // e with acute accent
			wantRune:  'e',
			wantCombc: []rune{'\u0301'},
			wantSize:  3,
		},
		{
			name:      "character with multiple combining marks",
			input:     []byte("e\u0301\u0308"), // e with acute and diaeresis
			wantRune:  'e',
			wantCombc: []rune{'\u0301', '\u0308'},
			wantSize:  5,
		},
		{
			name:      "emoji",
			input:     []byte("😀"),
			wantRune:  '😀',
			wantCombc: nil,
			wantSize:  4,
		},
		{
			name:      "empty byte array",
			input:     []byte{},
			wantRune:  utf8.RuneError,
			wantCombc: nil,
			wantSize:  0,
		},
		{
			name:      "single character only",
			input:     []byte("a"),
			wantRune:  'a',
			wantCombc: nil,
			wantSize:  1,
		},
		{
			name:      "precomposed acute e",
			input:     []byte("é"), // precomposed é
			wantRune:  'é',
			wantCombc: nil,
			wantSize:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRune, gotCombc, gotSize := DecodeCharacter(tt.input)
			if gotRune != tt.wantRune {
				t.Errorf("DecodeCharacter() rune = %q, want %q", gotRune, tt.wantRune)
			}
			if !reflect.DeepEqual(gotCombc, tt.wantCombc) {
				t.Errorf("DecodeCharacter() combc = %v, want %v", gotCombc, tt.wantCombc)
			}
			if gotSize != tt.wantSize {
				t.Errorf("DecodeCharacter() size = %v, want %v", gotSize, tt.wantSize)
			}
		})
	}
}

func TestDecodeCharacterInString(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantRune  rune
		wantCombc []rune
		wantSize  int
	}{
		{
			name:      "simple ASCII",
			input:     "hello",
			wantRune:  'h',
			wantCombc: nil,
			wantSize:  1,
		},
		{
			name:      "simple unicode",
			input:     "世界",
			wantRune:  '世',
			wantCombc: nil,
			wantSize:  3,
		},
		{
			name:      "character with combining accent",
			input:     "e\u0301", // e with acute accent
			wantRune:  'e',
			wantCombc: []rune{'\u0301'},
			wantSize:  3,
		},
		{
			name:      "character with multiple combining marks",
			input:     "e\u0301\u0308", // e with acute and diaeresis
			wantRune:  'e',
			wantCombc: []rune{'\u0301', '\u0308'},
			wantSize:  5,
		},
		{
			name:      "emoji",
			input:     "😀",
			wantRune:  '😀',
			wantCombc: nil,
			wantSize:  4,
		},
		{
			name:      "empty string",
			input:     "",
			wantRune:  utf8.RuneError,
			wantCombc: nil,
			wantSize:  0,
		},
		{
			name:      "single character only",
			input:     "a",
			wantRune:  'a',
			wantCombc: nil,
			wantSize:  1,
		},
		{
			name:      "precomposed acute e",
			input:     "é", // precomposed é
			wantRune:  'é',
			wantCombc: nil,
			wantSize:  2,
		},
		{
			name:      "combining grave accent",
			input:     "a\u0300", // a with grave
			wantRune:  'a',
			wantCombc: []rune{'\u0300'},
			wantSize:  3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRune, gotCombc, gotSize := DecodeCharacterInString(tt.input)
			if gotRune != tt.wantRune {
				t.Errorf("DecodeCharacterInString() rune = %q, want %q", gotRune, tt.wantRune)
			}
			if !reflect.DeepEqual(gotCombc, tt.wantCombc) {
				t.Errorf("DecodeCharacterInString() combc = %v, want %v", gotCombc, tt.wantCombc)
			}
			if gotSize != tt.wantSize {
				t.Errorf("DecodeCharacterInString() size = %v, want %v", gotSize, tt.wantSize)
			}
		})
	}
}

func TestCharacterCount(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  int
	}{
		{
			name:  "simple ASCII",
			input: []byte("hello"),
			want:  5,
		},
		{
			name:  "ASCII with spaces",
			input: []byte("hello world"),
			want:  11,
		},
		{
			name:  "empty byte array",
			input: []byte{},
			want:  0,
		},
		{
			name:  "single character",
			input: []byte("a"),
			want:  1,
		},
		{
			name:  "unicode characters",
			input: []byte("世界"),
			want:  2,
		},
		{
			name:  "mixed ASCII and unicode",
			input: []byte("hello 世界"),
			want:  8,
		},
		{
			name:  "character with combining accent",
			input: []byte("e\u0301"), // e with acute accent - should count as 1
			want:  1,
		},
		{
			name:  "multiple characters with combining marks",
			input: []byte("e\u0301a\u0300"), // é à - should count as 2
			want:  2,
		},
		{
			name:  "precomposed accented characters",
			input: []byte("café"),
			want:  4,
		},
		{
			name:  "emoji",
			input: []byte("hello 😀 world"),
			want:  13,
		},
		{
			name:  "multiple emojis",
			input: []byte("😀😁😂"),
			want:  3,
		},
		{
			name:  "newlines and tabs",
			input: []byte("hello\n\tworld"),
			want:  12,
		},
		{
			name:  "combining marks only (edge case)",
			input: []byte("\u0301\u0308"), // combining marks without base
			want:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CharacterCount(tt.input)
			if got != tt.want {
				t.Errorf("CharacterCount(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestCharacterCountInString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{
			name:  "simple ASCII",
			input: "hello",
			want:  5,
		},
		{
			name:  "ASCII with spaces",
			input: "hello world",
			want:  11,
		},
		{
			name:  "empty string",
			input: "",
			want:  0,
		},
		{
			name:  "single character",
			input: "a",
			want:  1,
		},
		{
			name:  "unicode characters",
			input: "世界",
			want:  2,
		},
		{
			name:  "mixed ASCII and unicode",
			input: "hello 世界",
			want:  8,
		},
		{
			name:  "character with combining accent",
			input: "e\u0301", // e with acute accent - should count as 1
			want:  1,
		},
		{
			name:  "multiple characters with combining marks",
			input: "e\u0301a\u0300", // é à - should count as 2
			want:  2,
		},
		{
			name:  "precomposed accented characters",
			input: "café",
			want:  4,
		},
		{
			name:  "emoji",
			input: "hello 😀 world",
			want:  13,
		},
		{
			name:  "multiple emojis",
			input: "😀😁😂",
			want:  3,
		},
		{
			name:  "newlines and tabs",
			input: "hello\n\tworld",
			want:  12,
		},
		{
			name:  "combining marks only (edge case)",
			input: "\u0301\u0308", // combining marks without base
			want:  0,
		},
		{
			name:  "complex combining sequence",
			input: "e\u0301\u0308\u0302", // e with multiple combining marks
			want:  1,
		},
		{
			name:  "multiple words with accents",
			input: "Montréal naïve café",
			want:  19,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CharacterCountInString(tt.input)
			if got != tt.want {
				t.Errorf("CharacterCountInString(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestDecodeCharacter_Sequential(t *testing.T) {
	// Test decoding multiple characters in sequence
	input := []byte("hello")
	expected := []rune{'h', 'e', 'l', 'l', 'o'}

	for i, expectedRune := range expected {
		r, combc, size := DecodeCharacter(input)
		if r != expectedRune {
			t.Errorf("Character %d: got %q, want %q", i, r, expectedRune)
		}
		if combc != nil {
			t.Errorf("Character %d: got combining marks %v, want nil", i, combc)
		}
		if size != 1 {
			t.Errorf("Character %d: got size %d, want 1", i, size)
		}
		input = input[size:]
	}
}

func TestDecodeCharacterInString_Sequential(t *testing.T) {
	// Test decoding multiple characters in sequence
	input := "world"
	expected := []rune{'w', 'o', 'r', 'l', 'd'}

	for i, expectedRune := range expected {
		r, combc, size := DecodeCharacterInString(input)
		if r != expectedRune {
			t.Errorf("Character %d: got %q, want %q", i, r, expectedRune)
		}
		if combc != nil {
			t.Errorf("Character %d: got combining marks %v, want nil", i, combc)
		}
		if size != 1 {
			t.Errorf("Character %d: got size %d, want 1", i, size)
		}
		input = input[size:]
	}
}

func TestCharacterCount_Consistency(t *testing.T) {
	// Test that CharacterCount and CharacterCountInString give the same results
	testStrings := []string{
		"hello",
		"世界",
		"hello 世界",
		"e\u0301",
		"café",
		"😀😁😂",
		"Montréal",
	}

	for _, str := range testStrings {
		byteCount := CharacterCount([]byte(str))
		stringCount := CharacterCountInString(str)
		if byteCount != stringCount {
			t.Errorf("Inconsistent counts for %q: CharacterCount=%d, CharacterCountInString=%d",
				str, byteCount, stringCount)
		}
	}
}

func BenchmarkCharacterCount(b *testing.B) {
	input := []byte("hello world this is a test string with some unicode 世界 characters")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CharacterCount(input)
	}
}

func BenchmarkCharacterCountInString(b *testing.B) {
	input := "hello world this is a test string with some unicode 世界 characters"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CharacterCountInString(input)
	}
}

func BenchmarkDecodeCharacter(b *testing.B) {
	input := []byte("hello world")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DecodeCharacter(input)
	}
}

func BenchmarkDecodeCharacterInString(b *testing.B) {
	input := "hello world"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DecodeCharacterInString(input)
	}
}
