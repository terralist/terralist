package cli

import (
	"os"
	"strings"
	"testing"
)

var _ Flag = (*IntFlag)(nil)

func TestIntFlag_IsHidden_False(t *testing.T) {
	flag := &IntFlag{}
	if flag.IsHidden() {
		t.Fatal("expected hidden false")
	}
}

func TestIntFlag_IsHidden_True(t *testing.T) {
	flag := &IntFlag{Hidden: true}
	if !flag.IsHidden() {
		t.Fatal("expected hidden true")
	}
}

func TestIntFlag_Set_NilUsesDefault(t *testing.T) {
	flag := IntFlag{DefaultValue: 7}
	if err := flag.Set(nil); err != nil {
		t.Fatal(err)
	}
	if flag.Value != 7 {
		t.Errorf("Value = %d", flag.Value)
	}
}

func TestIntFlag_Set_NilDoesNotMarkSet(t *testing.T) {
	flag := IntFlag{DefaultValue: 7}
	if err := flag.Set(nil); err != nil {
		t.Fatal(err)
	}
	if flag.IsSet() {
		t.Error("expected IsSet false")
	}
}

func TestIntFlag_Set_Int(t *testing.T) {
	flag := IntFlag{DefaultValue: 7}
	if err := flag.Set(9); err != nil {
		t.Fatal(err)
	}
	if flag.Value != 9 {
		t.Errorf("Value = %d", flag.Value)
	}
}

func TestIntFlag_Set_IntMarksSet(t *testing.T) {
	flag := IntFlag{DefaultValue: 7}
	if err := flag.Set(9); err != nil {
		t.Fatal(err)
	}
	if !flag.IsSet() {
		t.Error("expected IsSet true")
	}
}

func TestIntFlag_Set_ZeroYieldsZero(t *testing.T) {
	flag := IntFlag{DefaultValue: 7}
	if err := flag.Set(0); err != nil {
		t.Fatal(err)
	}
	if flag.Value != 0 {
		t.Errorf("Value = %d", flag.Value)
	}
}

func TestIntFlag_Set_ZeroMarksSetFalseWhenDifferentFromDefault(t *testing.T) {
	flag := IntFlag{DefaultValue: 7}
	if err := flag.Set(0); err != nil {
		t.Fatal(err)
	}
	if flag.IsSet() {
		t.Error("expected IsSet false for zero")
	}
}

func TestIntFlag_Set_ZeroMatchesDefault(t *testing.T) {
	flag := IntFlag{DefaultValue: 0}
	if err := flag.Set(0); err != nil {
		t.Fatal(err)
	}
	if flag.Value != 0 {
		t.Errorf("Value = %d", flag.Value)
	}
}

func TestIntFlag_Set_FromString(t *testing.T) {
	flag := IntFlag{}
	if err := flag.Set("42"); err != nil {
		t.Fatal(err)
	}
	if flag.Value != 42 {
		t.Errorf("Value = %d", flag.Value)
	}
}

func TestIntFlag_Set_FromStringMarksSet(t *testing.T) {
	flag := IntFlag{}
	if err := flag.Set("42"); err != nil {
		t.Fatal(err)
	}
	if !flag.IsSet() {
		t.Error("expected IsSet true")
	}
}

func TestIntFlag_Set_InvalidString(t *testing.T) {
	flag := IntFlag{}
	if err := flag.Set("nope"); err == nil {
		t.Fatal("expected error")
	}
}

func TestIntFlag_Set_WrongType(t *testing.T) {
	flag := IntFlag{}
	if err := flag.Set(true); err == nil {
		t.Fatal("expected error")
	}
}

func TestIntFlag_Set_Environment(t *testing.T) {
	t.Setenv("TL_INT_FLAG", "11")
	flag := IntFlag{}
	if err := flag.Set("${TL_INT_FLAG}"); err != nil {
		t.Fatal(err)
	}
	if flag.Value != 11 {
		t.Errorf("Value = %d", flag.Value)
	}
}

func TestIntFlag_Set_EnvironmentDefault(t *testing.T) {
	os.Unsetenv("TL_INT_FLAG_MISSING")
	flag := IntFlag{}
	if err := flag.Set("${TL_INT_FLAG_MISSING:5}"); err != nil {
		t.Fatal(err)
	}
	if flag.Value != 5 {
		t.Errorf("Value = %d", flag.Value)
	}
}

func TestIntFlag_Format_WithoutDefault(t *testing.T) {
	flag := IntFlag{Description: "port"}
	if flag.Format() != "port" {
		t.Errorf("Format() = %q", flag.Format())
	}
}

func TestIntFlag_Format_WithDefault(t *testing.T) {
	flag := IntFlag{Description: "port", DefaultValue: 8080}
	got := flag.Format()
	if !strings.Contains(got, "port") || !strings.Contains(got, "8080") {
		t.Errorf("Format() = %q", got)
	}
}

func TestIntFlag_Validate_RequiredNotSet(t *testing.T) {
	flag := IntFlag{Required: true}
	if err := flag.Validate(); err == nil {
		t.Fatal("expected required error")
	}
}

func TestIntFlag_Validate_RequiredSet(t *testing.T) {
	flag := IntFlag{Required: true}
	if err := flag.Set(1); err != nil {
		t.Fatal(err)
	}
	if err := flag.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestIntFlag_Validate_Optional(t *testing.T) {
	flag := IntFlag{}
	if err := flag.Validate(); err != nil {
		t.Fatal(err)
	}
}
