package cli

import (
	"os"
	"strings"
	"testing"
)

var _ Flag = (*BoolFlag)(nil)

func TestBoolFlag_IsHidden_False(t *testing.T) {
	flag := &BoolFlag{}
	if flag.IsHidden() {
		t.Fatal("expected hidden false")
	}
}

func TestBoolFlag_IsHidden_True(t *testing.T) {
	flag := &BoolFlag{Hidden: true}
	if !flag.IsHidden() {
		t.Fatal("expected hidden true")
	}
}

func TestBoolFlag_Set_NilUsesDefault(t *testing.T) {
	flag := BoolFlag{DefaultValue: true}
	if err := flag.Set(nil); err != nil {
		t.Fatal(err)
	}
	if flag.Value != true {
		t.Errorf("Value = %v", flag.Value)
	}
}

func TestBoolFlag_Set_NilDoesNotMarkSet(t *testing.T) {
	flag := BoolFlag{DefaultValue: true}
	if err := flag.Set(nil); err != nil {
		t.Fatal(err)
	}
	if flag.IsSet() {
		t.Error("expected IsSet false")
	}
}

func TestBoolFlag_Set_True(t *testing.T) {
	flag := BoolFlag{DefaultValue: false}
	if err := flag.Set(true); err != nil {
		t.Fatal(err)
	}
	if flag.Value != true {
		t.Errorf("Value = %v", flag.Value)
	}
}

func TestBoolFlag_Set_TrueMarksSet(t *testing.T) {
	flag := BoolFlag{DefaultValue: false}
	if err := flag.Set(true); err != nil {
		t.Fatal(err)
	}
	if !flag.IsSet() {
		t.Error("expected IsSet true")
	}
}

func TestBoolFlag_Set_FalseYieldsFalse(t *testing.T) {
	flag := BoolFlag{DefaultValue: true}
	if err := flag.Set(false); err != nil {
		t.Fatal(err)
	}
	if flag.Value != false {
		t.Errorf("Value = %v", flag.Value)
	}
}

func TestBoolFlag_Set_FalseMarksSetFalseWhenDifferentFromDefault(t *testing.T) {
	flag := BoolFlag{DefaultValue: true}
	if err := flag.Set(false); err != nil {
		t.Fatal(err)
	}
	if flag.IsSet() {
		t.Error("expected IsSet false for false value")
	}
}

func TestBoolFlag_Set_FalseMatchesDefault(t *testing.T) {
	flag := BoolFlag{DefaultValue: false}
	if err := flag.Set(false); err != nil {
		t.Fatal(err)
	}
	if flag.Value != false {
		t.Errorf("Value = %v", flag.Value)
	}
}

func TestBoolFlag_Set_FromString(t *testing.T) {
	flag := BoolFlag{}
	if err := flag.Set("true"); err != nil {
		t.Fatal(err)
	}
	if flag.Value != true {
		t.Errorf("Value = %v", flag.Value)
	}
}

func TestBoolFlag_Set_InvalidString(t *testing.T) {
	flag := BoolFlag{}
	if err := flag.Set("invalid"); err == nil {
		t.Fatal("expected error")
	}
}

func TestBoolFlag_Set_WrongType(t *testing.T) {
	flag := BoolFlag{}
	if err := flag.Set(1); err == nil {
		t.Fatal("expected error")
	}
}

func TestBoolFlag_Set_Environment(t *testing.T) {
	t.Setenv("TL_BOOL_FLAG", "true")
	flag := BoolFlag{}
	if err := flag.Set("${TL_BOOL_FLAG}"); err != nil {
		t.Fatal(err)
	}
	if flag.Value != true {
		t.Errorf("Value = %v", flag.Value)
	}
}

func TestBoolFlag_Set_EnvironmentDefault(t *testing.T) {
	os.Unsetenv("TL_BOOL_FLAG_MISSING")
	flag := BoolFlag{}
	if err := flag.Set("${TL_BOOL_FLAG_MISSING:true}"); err != nil {
		t.Fatal(err)
	}
	if flag.Value != true {
		t.Errorf("Value = %v", flag.Value)
	}
}

func TestBoolFlag_Format(t *testing.T) {
	flag := BoolFlag{Description: "debug", DefaultValue: true}
	got := flag.Format()
	if !strings.Contains(got, "debug") || !strings.Contains(got, "true") {
		t.Errorf("Format() = %q", got)
	}
}

func TestBoolFlag_Validate_RequiredNotSet(t *testing.T) {
	flag := BoolFlag{Required: true}
	if err := flag.Validate(); err == nil {
		t.Fatal("expected required error")
	}
}

func TestBoolFlag_Validate_RequiredSet(t *testing.T) {
	flag := BoolFlag{Required: true}
	if err := flag.Set(true); err != nil {
		t.Fatal(err)
	}
	if err := flag.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestBoolFlag_Validate_Optional(t *testing.T) {
	flag := BoolFlag{}
	if err := flag.Validate(); err != nil {
		t.Fatal(err)
	}
}
