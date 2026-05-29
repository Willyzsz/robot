package apperr

import (
	"errors"
	"reflect"
	"testing"
)

func TestFormatWalksWrappedErrorChain(t *testing.T) {
	root := errors.New("empty")
	err := Wrap("NewTeam", "name cannot be empty", root, Field{Name: "name", Value: ""})
	err = Wrap("CreateTeam", "create team", err)
	err = Wrap("HandlerCreateTeam", "handle create team", err, Field{Name: "request_id", Value: "abc"})

	got := Format(err)
	want := "[HandlerCreateTeam -> CreateTeam -> NewTeam]: {request_id=abc, name=}: handle create team: create team: name cannot be empty: empty"

	if got != want {
		t.Fatalf("Format() mismatch\nwant: %q\n got: %q", want, got)
	}
}

func TestWrapKeepsErrorsIsWorking(t *testing.T) {
	root := errors.New("not found")
	err := Wrap("FindByID", "", root, Field{Name: "id", Value: 10})
	err = Wrap("HandlerFindByID", "request failed", err)

	if !errors.Is(err, root) {
		t.Fatal("wrapped error should match root cause with errors.Is")
	}
}

func TestFieldsAndOps(t *testing.T) {
	err := Wrap("NewTeam", "name cannot be empty", errors.New("empty"), Field{Name: "name", Value: ""})
	err = Wrap("CreateTeam", "create team", err, Field{Name: "category_id", Value: 1})

	gotOps := Ops(err)
	wantOps := []string{"CreateTeam", "NewTeam"}
	if !reflect.DeepEqual(gotOps, wantOps) {
		t.Fatalf("Ops() mismatch\nwant: %#v\n got: %#v", wantOps, gotOps)
	}

	gotFields := Fields(err)
	wantFields := []Field{
		{Name: "category_id", Value: 1},
		{Name: "name", Value: ""},
	}
	if !reflect.DeepEqual(gotFields, wantFields) {
		t.Fatalf("Fields() mismatch\nwant: %#v\n got: %#v", wantFields, gotFields)
	}
}

func TestWrapNilErrorReturnsNil(t *testing.T) {
	if Wrap("AnyOp", "anything", nil) != nil {
		t.Fatal("Wrap() should return nil when err is nil")
	}
}
