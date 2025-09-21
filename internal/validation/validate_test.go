package validation

import "testing"

func TestNotBlank(t *testing.T) {
    if NotBlank("x", "value") != nil { t.Fatal("unexpected error") }
    if NotBlank("x", "   ") == nil { t.Fatal("expected error for blanks") }
}

func TestMinLen(t *testing.T) {
    if MinLen("p", "abcdef", 3) != nil { t.Fatal("unexpected error") }
    if MinLen("p", "a", 3) == nil { t.Fatal("expected error for short string") }
}
