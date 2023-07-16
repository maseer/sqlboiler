package drivers

import (
	"reflect"
	"testing"
)

func TestTablesFromList(t *testing.T) {
	t.Parallel()

	if TablesFromList(nil) != nil {
		t.Error("expected a shortcut to getting nil back")
	}

	if got := TablesFromList([]string{"a.b", "b", "c.d"}); !reflect.DeepEqual(got, []string{"b"}) {
		t.Error("list was wrong:", got)
	}
}

func TestColumnsFromList(t *testing.T) {
	t.Parallel()

	if ColumnsFromList(nil, "table") != nil {
		t.Error("expected a shortcut to getting nil back")
	}

	if got := ColumnsFromList([]string{"a.b", "b", "c.d", "c.a"}, "c"); !reflect.DeepEqual(got, []string{"d", "a"}) {
		t.Error("list was wrong:", got)
	}
	if got := ColumnsFromList([]string{"a.b", "b", "c.d", "c.a"}, "b"); len(got) != 0 {
		t.Error("list was wrong:", got)
	}
	if got := ColumnsFromList([]string{"*.b", "b", "c.d"}, "c"); !reflect.DeepEqual(got, []string{"b", "d"}) {
		t.Error("list was wrong:", got)
	}
}
