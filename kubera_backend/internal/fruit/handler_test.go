package fruit

import "testing"

func TestNormalize(t *testing.T) {
	tests := []struct{ name, unit, wantName, wantUnit, wantError string }{
		{"  Green   Apple  ", "KG", "Green Apple", "kg", ""},
		{"Mango", "", "Mango", "box", ""},
		{"", "kg", "", "", "name is required"},
		{"Orange", "bag", "", "", "default_unit must be box, kg, piece, crate, or dozen"},
	}
	for _, test := range tests {
		gotName, gotUnit, gotError := normalize(test.name, test.unit)
		if gotName != test.wantName || gotUnit != test.wantUnit || gotError != test.wantError {
			t.Fatalf("normalize(%q, %q) = %q, %q, %q", test.name, test.unit, gotName, gotUnit, gotError)
		}
	}
}
