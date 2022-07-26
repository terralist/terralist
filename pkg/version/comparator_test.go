package version

import "testing"

type compareTestData struct {
	left   string
	right  string
	expect int
}

var (
	compareTests = []compareTestData{
		{"1.0.0", "1.0.0", 0},
		{"1.1.0", "1.1.0", 0},
		{"1.1.1", "1.1.1", 0},
		{"1.1.1-pre", "1.1.1-pre", 0},
		{"1.1.1-pre+meta", "1.1.1-pre+meta", 0},
		{"1.0.0", "1.0.1", -1},
		{"1.0.0", "1.1.0", -1},
		{"1.1.0", "1.0.0", 1},
		{"2.0.0", "1.0.0", 1},
		{"2.0.1", "2.0.0", 1},
		{"2.0.0-alpha", "2.0.0-beta", -1},
		{"2.0.0-beta", "2.0.0-alpha", 1},
		{"2.0.0-pre+alpha", "2.0.0-pre+beta", -1},
		{"2.0.0-pre+beta", "2.0.0-pre+alpha", 1},
	}
)

func TestCompare(t *testing.T) {
	for _, test := range compareTests {
		got := Compare(Version(test.left), Version(test.right))
		want := test.expect

		if got != want {
			t.Errorf("Comparing %v with %v: got %v, expecting %v.", test.left, test.right, got, want)
		}
	}
}
