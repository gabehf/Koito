package romanizer

import "testing"

// --------------------
// Basic / Latin Tests
// --------------------

func TestRomanize_Basic(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"   ", ""},
		{"Hello World", ""},
		{"Hello, World!", ""},
		{"CSCI 630 - Koito", ""},
	}

	for _, tt := range tests {
		result := Romanize(tt.input)
		if result != tt.expected {
			t.Errorf("Romanize(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

// --------------------
// Japanese
// --------------------

func TestRomanize_Japanese(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"こんにちは", "konnichiha"},
		{"   こんにちは   ", "konnichiha"},
	}

	for _, tt := range tests {
		result := Romanize(tt.input)
		if result != tt.expected {
			t.Errorf("Romanize(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

// --------------------
// Korean
// --------------------

func TestRomanize_Korean(t *testing.T) {
	result := Romanize("안녕하세요")
	expected := "annyeonghaseyo"

	if result != expected {
		t.Errorf("Romanize(%q) = %q, expected %q", "안녕하세요", result, expected)
	}
}

// --------------------
// Chinese
// --------------------

func TestRomanize_Chinese(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"你好", "Ni Hao"},
		{"Hello 世界", "Hello Shi Jie"},
	}

	for _, tt := range tests {
		result := Romanize(tt.input)
		if result != tt.expected {
			t.Errorf("Romanize(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

// --------------------
// Russian (Cyrillic)
// --------------------

func TestRomanize_Russian(t *testing.T) {
	result := Romanize("Привет мир")
	expected := "Privet mir"

	if result != expected {
		t.Errorf("Romanize(%q) = %q, expected %q", "Привет мир", result, expected)
	}
}

// --------------------
// Greek
// --------------------

func TestRomanize_Greek(t *testing.T) {
	result := Romanize("Γειά σου Κόσμε")
	expected := "Geia sou Kosme"

	if result != expected {
		t.Errorf("Romanize(%q) = %q, expected %q", "Γειά σου Κόσμε", result, expected)
	}
}

// --------------------
// Arabic
// --------------------

func TestRomanize_Arabic(t *testing.T) {
	result := Romanize("مرحبا بالعالم")
	expected := "mrHb bl`lm"

	if result != expected {
		t.Errorf("Romanize(%q) = %q, expected %q", "مرحبا بالعالم", result, expected)
	}
}

// --------------------
// Hebrew
// --------------------

func TestRomanize_Hebrew(t *testing.T) {
	result := Romanize("שלום עולם")
	expected := "shlvm `vlm"

	if result != expected {
		t.Errorf("Romanize(%q) = %q, expected %q", "שלום עולם", result, expected)
	}
}

// --------------------
// Hindi (Devanagari)
// --------------------

func TestRomanize_Hindi(t *testing.T) {
	result := Romanize("नमस्ते")
	expected := "nmste"

	if result != expected {
		t.Errorf("Romanize(%q) = %q, expected %q", "नमस्ते", result, expected)
	}
}

// --------------------
// Thai
// --------------------

func TestRomanize_Thai(t *testing.T) {
	result := Romanize("สวัสดี")
	expected := "swasdii"

	if result != expected {
		t.Errorf("Romanize(%q) = %q, expected %q", "สวัสดี", result, expected)
	}
}

// --------------------
// Emoji / Symbols
// --------------------

func TestRomanize_Emoji(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"🙂", ""},
		{"Hello 🙂", "Hello"},
	}

	for _, tt := range tests {
		result := Romanize(tt.input)
		if result != tt.expected {
			t.Errorf("Romanize(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}
