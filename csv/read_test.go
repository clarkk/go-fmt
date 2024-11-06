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

func Test_error(t *testing.T){
	t.Run("empty", func(t *testing.T){
		tests := []test_error{{
			reader:	func(t *testing.T) *Reader {
				return NewReader("")
			},
			input:	"",
			error:	"CSV empty",
		},{
			reader:	func(t *testing.T) *Reader {
				return NewReader("")
			},
			input:	"head1\ntest1",
			error:	"CSV must have more than one column",
		},{
			reader:	func(t *testing.T) *Reader {
				return NewReader("")
			},
			input:	"head1,head2",
			error:	"CSV empty",
		},{
			reader:	func(t *testing.T) *Reader {
				return NewReader("").
					Remove_overflow_cols()
			},
			input:	"head1,head2",
			error:	"CSV empty",
		}}
		verify_test(t, tests)
	})
	
	t.Run("too few column headers", func(t *testing.T){
		tests := []test_error{{
			reader:	func(t *testing.T) *Reader {
				return NewReader("")
			},
			input:	"head1\ntest1,test2",
			error:	"CSV has too few column headers",
		},{
			reader:	func(t *testing.T) *Reader {
				return NewReader("").
					Remove_overflow_cols()
			},
			input:	"100\ntest1,test2",
			error:	"CSV has too few column headers",
		}}
		verify_test(t, tests)
	})
	
	t.Run("columns not equal", func(t *testing.T){
		tests := []test_error{{
			reader:	func(t *testing.T) *Reader {
				return NewReader("").
					Col_integrity()
			},
			input:	"head1,head2\ntest1",
			error:	"Columns in CSV not equal",
		}}
		verify_test(t, tests)
	})
	
	t.Run("column headers", func(t *testing.T){
		tests := []test_error{{
			reader:	func(t *testing.T) *Reader {
				return NewReader("")
			},
			input:	"head1,,head3\ntest1,test2,test3\ntest1,test2,test3",
			error:	"Column headers cannot be empty",
		},{
			reader:	func(t *testing.T) *Reader {
				return NewReader("")
			},
			input:	"head1,100,head3\ntest1,test2,test3\ntest1,test2,test3",
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
			input:	"test1,test2,test3\ntest1\ntest1,test2",
			rows:	"test1,test2,test3\ntest1,,\ntest1,test2,",
		},{
			reader:	func(t *testing.T) *Reader {
				return NewReader("").
					Ignore_header()
			},
			input:	"test1,test2\ntest1\ntest1,test2,test3",
			rows:	"test1,test2,\ntest1,,\ntest1,test2,test3",
		}}
		verify_test(t, tests)
	})
	
	t.Run("remove empty columns", func(t *testing.T){
		tests := []test_output{{
			reader:	func(t *testing.T) *Reader {
				return NewReader("").
					Remove_empty_cols()
			},
			input:	"head1,head2,head3\ntest1,,test3\ntest1,,test3",
			header:	"head1,head3",
			rows:	"test1,test3\ntest1,test3",
		},{
			reader:	func(t *testing.T) *Reader {
				return NewReader("").
					Remove_empty_cols()
			},
			input:	"head1,,head3\ntest1,,test3\ntest1,,test3",
			header:	"head1,head3",
			rows:	"test1,test3\ntest1,test3",
		}}
		verify_test(t, tests)
	})
	
	t.Run("header and rows", func(t *testing.T){
		tests := []test_output{{
			reader:	func(t *testing.T) *Reader {
				return NewReader("")
			},
			input:	"head1,head2,head3\ntest1,test2,test3\ntest1,test2,test3",
			header:	"head1,head2,head3",
			rows:	"test1,test2,test3\ntest1,test2,test3",
		},{
			reader:	func(t *testing.T) *Reader {
				return NewReader("").
					Optional_header()
			},
			input:	"head1,head2,head3\ntest1,test2,test3\ntest1,test2,test3",
			header:	"head1,head2,head3",
			rows:	"test1,test2,test3\ntest1,test2,test3",
		},{
			reader:	func(t *testing.T) *Reader {
				return NewReader("").
					Optional_header()
			},
			input:	"test1,,test3\ntest1,test2,test3",
			rows:	"test1,,test3\ntest1,test2,test3",
		},{
			reader:	func(t *testing.T) *Reader {
				return NewReader("").
					Ignore_header()
			},
			input:	"head1,head2,head3\ntest1,test2,test3\ntest1,test2,test3",
			rows:	"head1,head2,head3\ntest1,test2,test3\ntest1,test2,test3",
		},{
			reader:	func(t *testing.T) *Reader {
				return NewReader("").
					Remove_overflow_cols()
			},
			input:	"head1,head2,head3\ntest1,test2,test3,test4\ntest1,test2",
			header:	"head1,head2,head3",
			rows:	"test1,test2,test3\ntest1,test2,",
		},{
			reader:	func(t *testing.T) *Reader {
				return NewReader("").
					Remove_empty_cols().
					Remove_overflow_cols()
			},
			input:	"head1,head2,,head4\ntest1,test2,,test4,test5\ntest1,",
			header:	"head1,head2,head4",
			rows:	"test1,test2,test4\ntest1,,",
		},{
			reader:	func(t *testing.T) *Reader {
				return NewReader("").
					Remove_empty_cols().
					Remove_overflow_cols()
			},
			input:	"head1,head2,head3,head4\ntest1,test2,,test4,test5\ntest1,",
			header:	"head1,head2,head4",
			rows:	"test1,test2,test4\ntest1,,",
		}}
		verify_test(t, tests)
	})
}

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