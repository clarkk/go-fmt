package csv

import (
	"fmt"
	"strings"
	"testing"
)

type (
	tester interface {
		verify(*testing.T)
	}
	
	test_error struct {
		reader		func(t *testing.T) *Reader
		input		string
		error		string
	}
	
	test_output struct {
		reader		func(t *testing.T) *Reader
		input		string
		header		string
		rows		string
	}
)

func (e test_error) verify(t *testing.T){
	r := e.reader(t)
	_, err := r.Bytes([]byte(e.input), "")
	if err == nil {
		t.Fatal("Expected an error")
	}
	
	if err.Error() != e.error {
		t.Fatalf("Expected error '%s', got '%v'", e.error, err)
	}
	
	fmt.Println(strings.Join(r.Log(), "\n"))
}

func (o test_output) verify(t *testing.T){
	r := o.reader(t)
	out, err := r.Bytes([]byte(o.input), "")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	
	header := strings.Join(out.Header, ",")
	if header != o.header {
		t.Fatalf("Want: %s\n\nGot: %s", o.header, header)
	}
	
	s := make([]string, len(out.Rows))
	for i, line := range out.Rows {
		s[i] = strings.Join(line.Row, ",")
	}
	
	rows := strings.Join(s, "\n")
	if rows != o.rows {
		t.Fatalf("Want: %s\n\nGot: %s", o.rows, rows)
	}
	
	fmt.Println(strings.Join(r.Log(), "\n"))
}

func verify_test[T tester](t *testing.T, tests []T){
	for i, tt := range tests {
		fmt.Println("test:", i, "\n---")
		tt.verify(t)
	}
}

func Test_error(t *testing.T){
	t.Run("unable to parse CSV", func(t *testing.T){
		tests := []test_error{{
			reader:	func(t *testing.T) *Reader {
				return NewReader("")
			},
			input:	`"test","`,
			error:	"Unable to parse CSV",
		}}
		verify_test(t, tests)
	})
	
	t.Run("empty CSV", func(t *testing.T){
		tests := []test_error{{
			reader:	func(t *testing.T) *Reader {
				return NewReader("")
			},
			input:	``,
			error:	"CSV empty",
		}}
		verify_test(t, tests)
	})
	
	t.Run("too few column headers", func(t *testing.T){
		tests := []test_error{{
			reader:	func(t *testing.T) *Reader {
				return NewReader("")
			},
			input:	"test\ntest,test",
			error:	"Too few column headers",
		}}
		verify_test(t, tests)
	})
	
	t.Run("columns in CSV not equal", func(t *testing.T){
		tests := []test_error{{
			reader:	func(t *testing.T) *Reader {
				return NewReader("").
					Col_integrity()
			},
			input:	"test,test\ntest",
			error:	"Columns in CSV not equal",
		}}
		verify_test(t, tests)
	})
	
	t.Run("column headers cannot be empty", func(t *testing.T){
		tests := []test_error{{
			reader:	func(t *testing.T) *Reader {
				return NewReader("")
			},
			input:	"head,,head\ntest,test,test\ntest,test,test",
			error:	"Column headers cannot be empty",
		}}
		verify_test(t, tests)
	})
	
	t.Run("column headers required", func(t *testing.T){
		tests := []test_error{{
			reader:	func(t *testing.T) *Reader {
				return NewReader("")
			},
			input:	"head,100,head\ntest,test,test\ntest,test,test",
			error:	"Column headers in CSV required",
		}}
		verify_test(t, tests)
	})
}

func Test_ouput(t *testing.T){
	t.Run("fill empty columns", func(t *testing.T){
		tests := []test_output{{
			reader:	func(t *testing.T) *Reader {
				return NewReader("").
					Ignore_header()
			},
			input:	"test,test,test\ntest\ntest,test",
			rows:	"test,test,test\ntest,,\ntest,test,",
		},{
			reader:	func(t *testing.T) *Reader {
				return NewReader("").
					Ignore_header()
			},
			input:	"test,test\ntest\ntest,test,test",
			rows:	"test,test,\ntest,,\ntest,test,test",
		}}
		verify_test(t, tests)
	})
	
	t.Run("remove empty columns", func(t *testing.T){
		tests := []test_output{{
			reader:	func(t *testing.T) *Reader {
				return NewReader("").
					Remove_empty_cols()
			},
			input:	"head,head,head\ntest,,test\ntest,,test",
			header:	"head,head",
			rows:	"test,test\ntest,test",
		},{
			reader:	func(t *testing.T) *Reader {
				return NewReader("").
					Remove_empty_cols()
			},
			input:	"head,,head\ntest,,test\ntest,,test",
			header:	"head,head",
			rows:	"test,test\ntest,test",
		}}
		verify_test(t, tests)
	})
	
	t.Run("header and rows", func(t *testing.T){
		tests := []test_output{{
			reader:	func(t *testing.T) *Reader {
				return NewReader("")
			},
			input:	"head,head,head\ntest,test,test\ntest,test,test",
			header:	"head,head,head",
			rows:	"test,test,test\ntest,test,test",
		},{
			reader:	func(t *testing.T) *Reader {
				return NewReader("").
					Optional_header()
			},
			input:	"head,head,head\ntest,test,test\ntest,test,test",
			header:	"head,head,head",
			rows:	"test,test,test\ntest,test,test",
		},{
			reader:	func(t *testing.T) *Reader {
				return NewReader("").
					Optional_header()
			},
			input:	"test,,test\ntest,test,test",
			rows:	"test,,test\ntest,test,test",
		},{
			reader:	func(t *testing.T) *Reader {
				return NewReader("").
					Ignore_header()
			},
			input:	"head,head,head\ntest,test,test\ntest,test,test",
			rows:	"head,head,head\ntest,test,test\ntest,test,test",
		}}
		verify_test(t, tests)
	})
}