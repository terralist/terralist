package cli

import (
	"os"
	"strings"
	"testing"
)

var _ Flag = (*StringFlag)(nil)

func TestStringFlag_IsHidden_False(t *testing.T) {
	flag := &StringFlag{}
	if flag.IsHidden() {
		t.Fatal("expected hidden false")
	}
}

func TestStringFlag_IsHidden_True(t *testing.T) {
	flag := &StringFlag{Hidden: true}
	if !flag.IsHidden() {
		t.Fatal("expected hidden true")
	}
}

func TestStringFlag_Set_NilUsesDefault(t *testing.T) {
	flag := StringFlag{DefaultValue: "fallback"}
	if err := flag.Set(nil); err != nil {
		t.Fatal(err)
	}
	if flag.Value != "fallback" {
		t.Errorf("Value = %q", flag.Value)
	}
}

func TestStringFlag_Set_NilDoesNotMarkSet(t *testing.T) {
	flag := StringFlag{DefaultValue: "fallback"}
	if err := flag.Set(nil); err != nil {
		t.Fatal(err)
	}
	if flag.IsSet() {
		t.Error("expected IsSet false")
	}
}

func TestStringFlag_Set_String(t *testing.T) {
	flag := StringFlag{DefaultValue: "fallback"}
	if err := flag.Set("hello"); err != nil {
		t.Fatal(err)
	}
	if flag.Value != "hello" {
		t.Errorf("Value = %q", flag.Value)
	}
}

func TestStringFlag_Set_StringMarksSet(t *testing.T) {
	flag := StringFlag{DefaultValue: "fallback"}
	if err := flag.Set("hello"); err != nil {
		t.Fatal(err)
	}
	if !flag.IsSet() {
		t.Error("expected IsSet true")
	}
}

func TestStringFlag_Set_EmptyYieldsEmpty(t *testing.T) {
	flag := StringFlag{DefaultValue: "fallback"}
	if err := flag.Set(""); err != nil {
		t.Fatal(err)
	}
	if flag.Value != "" {
		t.Errorf("Value = %q, want empty", flag.Value)
	}
}

func TestStringFlag_Set_EmptyMarksSetFalseWhenDifferentFromDefault(t *testing.T) {
	flag := StringFlag{DefaultValue: "fallback"}
	if err := flag.Set(""); err != nil {
		t.Fatal(err)
	}
	if flag.IsSet() {
		t.Error("expected IsSet false for empty string")
	}
}

func TestStringFlag_Set_EmptyMatchesDefault(t *testing.T) {
	flag := StringFlag{DefaultValue: ""}
	if err := flag.Set(""); err != nil {
		t.Fatal(err)
	}
	if flag.Value != "" {
		t.Errorf("Value = %q", flag.Value)
	}
}

func TestStringFlag_Set_WrongType(t *testing.T) {
	flag := StringFlag{}
	if err := flag.Set(42); err == nil {
		t.Fatal("expected error")
	}
}

func TestStringFlag_Set_ValidChoice(t *testing.T) {
	flag := StringFlag{Choices: []string{"a", "b"}}
	if err := flag.Set("a"); err != nil {
		t.Fatal(err)
	}
}

func TestStringFlag_Set_InvalidChoice(t *testing.T) {
	flag := StringFlag{Choices: []string{"a", "b"}}
	if err := flag.Set("c"); err == nil {
		t.Fatal("expected error for invalid choice")
	}
}

func TestStringFlag_Set_Environment(t *testing.T) {
	t.Setenv("TL_STRING_FLAG", "from-env")
	flag := StringFlag{}
	if err := flag.Set("${TL_STRING_FLAG}"); err != nil {
		t.Fatal(err)
	}
	if flag.Value != "from-env" {
		t.Errorf("Value = %q", flag.Value)
	}
}

func TestStringFlag_Set_EnvironmentDefault(t *testing.T) {
	os.Unsetenv("TL_STRING_FLAG_MISSING")
	flag := StringFlag{}
	if err := flag.Set("${TL_STRING_FLAG_MISSING:def}"); err != nil {
		t.Fatal(err)
	}
	if flag.Value != "def" {
		t.Errorf("Value = %q", flag.Value)
	}
}

func TestStringFlag_Format_Description(t *testing.T) {
	flag := StringFlag{Description: "mode"}
	if !strings.Contains(flag.Format(), "mode") {
		t.Errorf("Format() = %q", flag.Format())
	}
}

func TestStringFlag_Format_Choices(t *testing.T) {
	flag := StringFlag{Description: "mode", Choices: []string{"a", "b"}}
	got := flag.Format()
	if !strings.Contains(got, "Options:") || !strings.Contains(got, "a, b") {
		t.Errorf("Format() = %q", got)
	}
}

func TestStringFlag_Format_Default(t *testing.T) {
	flag := StringFlag{Description: "mode", DefaultValue: "a"}
	if !strings.Contains(flag.Format(), `default "a"`) {
		t.Errorf("Format() = %q", flag.Format())
	}
}

func TestStringFlag_Validate_RequiredNotSet(t *testing.T) {
	flag := StringFlag{Required: true}
	if err := flag.Validate(); err == nil {
		t.Fatal("expected required error")
	}
}

func TestStringFlag_Validate_RequiredSet(t *testing.T) {
	flag := StringFlag{Required: true}
	if err := flag.Set("ok"); err != nil {
		t.Fatal(err)
	}
	if err := flag.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestStringFlag_Validate_OptionalEmpty(t *testing.T) {
	flag := StringFlag{}
	if err := flag.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestStringFlag_Validate_OptionalEmptyChoice(t *testing.T) {
	flag := StringFlag{Choices: []string{"a", "b"}}
	if err := flag.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestStringFlag_Validate_InvalidChoice(t *testing.T) {
	flag := StringFlag{Choices: []string{"a", "b"}, Value: "z"}
	if err := flag.Validate(); err == nil {
		t.Fatal("expected invalid choice")
	}
}

func TestStringFlag_Validate_ValidChoice(t *testing.T) {
	flag := StringFlag{Choices: []string{"a", "b"}, Value: "a"}
	if err := flag.Validate(); err != nil {
		t.Fatal(err)
	}
}
