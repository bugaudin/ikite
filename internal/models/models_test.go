package models

import "testing"

func TestCardinalDirection(t *testing.T) {
	cases := map[float64]string{
		0:   "N",
		45:  "NE",
		90:  "E",
		180: "S",
		270: "W",
		360: "N",
	}
	for angle, want := range cases {
		if got := CardinalDirection(angle); got != want {
			t.Fatalf("CardinalDirection(%v)=%q want %q", angle, got, want)
		}
	}
}

func TestWindCSSClass(t *testing.T) {
	cases := map[float64]string{
		7:  "",
		8:  "wind8",
		10: "wind10",
		14: "wind14",
		16: "wind16",
		18: "wind18",
		25: "wind18",
	}
	for wind, want := range cases {
		if got := WindCSSClass(wind); got != want {
			t.Fatalf("WindCSSClass(%v)=%q want %q", wind, got, want)
		}
	}
}
