package appenv

import "testing"

func mapToLookupFunc(m map[string]string) lookupFunc {
	return func(s string) (string, bool) {
		v, ok := m[s]
		return v, ok
	}
}

func testSetFields[T comparable](t *testing.T, env map[string]string, expect T) {
	lookup := mapToLookupFunc(env)
	var v T
	err := setFields(&v, lookup)
	if err != nil {
		t.Fatal(err)
	}
	if v != expect {
		t.Fatalf("expect %+v but got %+v", expect, v)
	}
}

func TestSetFields(t *testing.T) {
	t.Run("empty struct", func(t *testing.T) {
		testSetFields[struct{}](t, map[string]string{}, struct{}{})
	})

	t.Run("only string", func(t *testing.T) {
		type ABC struct {
			A string `env:"A"`
			B string `env:"B"`
			C string `env:"C"`
		}
		testSetFields[ABC](t, map[string]string{"A": "A", "B": "B", "C": "C"}, ABC{"A", "B", "C"})
	})

	t.Run("string, int, and bool", func(t *testing.T) {
		type S struct {
			Str  string `env:"STR"`
			Int  int    `env:"INT"`
			Bool bool   `env:"BOOL"`
		}
		testSetFields[S](t, map[string]string{
			"STR":  "string",
			"INT":  "123",
			"BOOL": "true",
		}, S{"string", 123, true})
	})

	t.Run("include irrelevant filed", func(t *testing.T) {
		type T struct {
			Str        string   `env:"STR"`
			Irr        struct{} `json:"irr"`
			Irr2       uint
			unexported string `env:"UNEXPORTED"`
			Int        int    `env:"INT"`
		}
		testSetFields[T](t, map[string]string{
			"STR":        "にゃんぱす",
			"irr":        "irr",
			"UNEXPORTED": "uoo",
			"INT":        "17",
		}, T{"にゃんぱす", struct{}{}, 0, "", 17})
	})

}

func TestLoad(t *testing.T) {
	t.Run("empty struct", func(t *testing.T) {
	})
}
