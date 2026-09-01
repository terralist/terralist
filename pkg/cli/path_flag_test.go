package cli

import (
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"testing"
)

var _ Flag = (*PathFlag)(nil)

func TestPathFlag_IsHidden_False(t *testing.T) {
	flag := &PathFlag{}
	if flag.IsHidden() {
		t.Fatal("expected hidden false")
	}
}

func TestPathFlag_IsHidden_True(t *testing.T) {
	flag := &PathFlag{Hidden: true}
	if !flag.IsHidden() {
		t.Fatal("expected hidden true")
	}
}

func TestPathFlag_Set_NilUsesDefault(t *testing.T) {
	flag := PathFlag{DefaultValue: "/tmp"}
	if err := flag.Set(nil); err != nil {
		t.Fatal(err)
	}
	if flag.Value != "/tmp" {
		t.Errorf("Value = %q", flag.Value)
	}
}

func TestPathFlag_Set_NilDoesNotMarkSet(t *testing.T) {
	flag := PathFlag{DefaultValue: "/tmp"}
	if err := flag.Set(nil); err != nil {
		t.Fatal(err)
	}
	if flag.IsSet() {
		t.Error("expected IsSet false")
	}
}

func TestPathFlag_Set_Absolute(t *testing.T) {
	flag := PathFlag{}
	if err := flag.Set("/var/lib/data"); err != nil {
		t.Fatal(err)
	}
	if flag.Value != "/var/lib/data" {
		t.Errorf("Value = %q", flag.Value)
	}
}

func TestPathFlag_Set_AbsoluteMarksSet(t *testing.T) {
	flag := PathFlag{}
	if err := flag.Set("/var/lib/data"); err != nil {
		t.Fatal(err)
	}
	if !flag.IsSet() {
		t.Error("expected IsSet true")
	}
}

func TestPathFlag_Set_EmptyYieldsEmpty(t *testing.T) {
	flag := PathFlag{DefaultValue: "/tmp"}
	if err := flag.Set(""); err != nil {
		t.Fatal(err)
	}
	if flag.Value != "" {
		t.Errorf("Value = %q", flag.Value)
	}
}

func TestPathFlag_Set_EmptyMarksSetFalseWhenDifferentFromDefault(t *testing.T) {
	flag := PathFlag{DefaultValue: "/tmp"}
	if err := flag.Set(""); err != nil {
		t.Fatal(err)
	}
	if flag.IsSet() {
		t.Error("expected IsSet false for empty path")
	}
}

func TestPathFlag_Set_WrongType(t *testing.T) {
	flag := PathFlag{}
	if err := flag.Set(1); err == nil {
		t.Fatal("expected error")
	}
}

func TestPathFlag_Set_HomeTilde(t *testing.T) {
	usr, err := user.Current()
	if err != nil {
		t.Fatal(err)
	}
	flag := PathFlag{}
	if err := flag.Set("~"); err != nil {
		t.Fatal(err)
	}
	if flag.Value != usr.HomeDir {
		t.Errorf("Value = %q, want %q", flag.Value, usr.HomeDir)
	}
}

func TestPathFlag_Set_HomeTildePath(t *testing.T) {
	usr, err := user.Current()
	if err != nil {
		t.Fatal(err)
	}
	flag := PathFlag{}
	if err := flag.Set("~/data"); err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(usr.HomeDir, "data")
	if flag.Value != want {
		t.Errorf("Value = %q, want %q", flag.Value, want)
	}
}

func TestPathFlag_Set_Dot(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	flag := PathFlag{}
	if err := flag.Set("."); err != nil {
		t.Fatal(err)
	}
	if flag.Value != cwd {
		t.Errorf("Value = %q, want %q", flag.Value, cwd)
	}
}

func TestPathFlag_Set_DotSlash(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	flag := PathFlag{}
	if err := flag.Set("./subdir"); err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(cwd, "subdir")
	if flag.Value != want {
		t.Errorf("Value = %q, want %q", flag.Value, want)
	}
}

func TestPathFlag_Set_Environment(t *testing.T) {
	t.Setenv("TL_PATH_FLAG", "/opt/terralist")
	flag := PathFlag{}
	if err := flag.Set("${TL_PATH_FLAG}"); err != nil {
		t.Fatal(err)
	}
	if flag.Value != "/opt/terralist" {
		t.Errorf("Value = %q", flag.Value)
	}
}

func TestPathFlag_Format(t *testing.T) {
	flag := PathFlag{Description: "config", DefaultValue: "/etc/app"}
	got := flag.Format()
	if !strings.Contains(got, "config") || !strings.Contains(got, "/etc/app") {
		t.Errorf("Format() = %q", got)
	}
}

func TestPathFlag_Validate_RequiredNotSet(t *testing.T) {
	flag := PathFlag{Required: true}
	if err := flag.Validate(); err == nil {
		t.Fatal("expected required error")
	}
}

func TestPathFlag_Validate_RequiredSet(t *testing.T) {
	flag := PathFlag{Required: true}
	if err := flag.Set("/abs"); err != nil {
		t.Fatal(err)
	}
	if err := flag.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestPathFlag_Validate_OptionalEmpty(t *testing.T) {
	flag := PathFlag{}
	if err := flag.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestPathFlag_Validate_Relative(t *testing.T) {
	flag := PathFlag{Value: "rel"}
	if err := flag.Validate(); err == nil {
		t.Fatal("expected non-absolute error")
	}
}
