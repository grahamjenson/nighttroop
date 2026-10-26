package css

import "testing"

func TestColour(t *testing.T) {
	// Tailwind's own hex for its oklch palette, and the brand colours as the theme has them
	tests := map[string]string{
		"brand-night": "#10203F",
		"white":       "#fff",
		"gray-100":    "#f3f4f6",
		"gray-200":    "#e5e7eb",
		"gray-700":    "#364153",
		"red-600":     "#e7000b",
		"amber-500":   "#fe9a00",
		"blue-500":    "#2b7fff",
	}
	for name, want := range tests {
		t.Run(name, func(t *testing.T) {
			if got := Colour(name); got != want {
				t.Errorf("Colour(%q) = %s, want %s", name, got, want)
			}
		})
	}
}

func TestColourPanicsForAColourNotInTheTheme(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("Colour(\"purple-500\") did not panic")
		}
	}()
	Colour("purple-500")
}
