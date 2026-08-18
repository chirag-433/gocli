package main

import "testing"

func TestCleanInput(t *testing.T) {
	cases := []struct {
		input    string
		expected []string
	}{
		{
			input: "Hellow WORLD",
			expected: []string{
				"hellow",

				"world",
			},
		},
	}
	for _, cs := range cases {
		actual := cleanInput(cs.input)
		if len(actual) != len(cs.expected) {
			t.Errorf("The lenghts are not equal: %v and %v", len(actual), len(cs.expected))
			continue
		}
		for i := range actual {
			actualWord := actual[i]
			expectedword := cs.expected[i]
			if actualWord != expectedword {
				t.Errorf("%v does now equal %v", actualWord, expectedword)

			}
		}
	}
}
