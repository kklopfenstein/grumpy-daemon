package utils

import "testing"

func TestContainsSearch(t *testing.T) {
	if !ContainsSearch("something else", "else") {
		t.Error(`ContainsSearch("something else", "else") = false`)
	}

	if !ContainsSearch("something else again", "else") {
		t.Error(`ContainsSearch("something else again", "else") = false`)
	}

	if ContainsSearch("something else again", "else+") {
		t.Error(`ContainsSearch("something else again", "else+") = true`)
	}

	if !ContainsSearch("else", "else") {
		t.Error(`ContainsSearch("else", "else") = false`)
	}

	if ContainsSearch("something else", "thing") {
		t.Error(`ContainsSearch("soemthing else", "thing") = true`)
	}
}
