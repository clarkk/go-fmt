package csv

import (
	"fmt"
	"strings"
	"testing"
)

func Test_reader(t *testing.T){
	t.Run("invalid file encoding", func(t *testing.T){
		eur_sign 		:= "\x80"
		s 				:= fmt.Sprintf(`"test","%s"`, eur_sign)
		err_expected	:= "Invalid file encoding"
		
		r := NewReader("")
		_, err := r.Bytes([]byte(s), "")
		if err == nil {
			t.Fatal("Expected an error")
		}
		
		if err.Error() != err_expected {
			t.Fatalf("Expected error '%s', got '%v'", err_expected, err)
		}
		
		fmt.Println(strings.Join(r.Log(), "\n"))
	})
	
	t.Run("unable to parse CSV", func(t *testing.T){
		s 				:= `"test","`
		err_expected	:= "Unable to parse CSV"
		
		r := NewReader("")
		_, err := r.Bytes([]byte(s), "")
		if err == nil {
			t.Fatal("Expected an error")
		}
		
		if err.Error() != err_expected {
			t.Fatalf("Expected error '%s', got '%v'", err_expected, err)
		}
		
		fmt.Println(strings.Join(r.Log(), "\n"))
	})
	
	t.Run("empty CSV", func(t *testing.T){
		s 				:= ""
		err_expected	:= "CSV empty"
		
		r := NewReader("")
		_, err := r.Bytes([]byte(s), "")
		if err == nil {
			t.Fatal("Expected an error")
		}
		
		if err.Error() != err_expected {
			t.Fatalf("Expected error '%s', got '%v'", err_expected, err)
		}
		
		fmt.Println(strings.Join(r.Log(), "\n"))
	})
	
	t.Run("columns in CSV not equal", func(t *testing.T){
		s 				:= `"test","test"`+"\n"+`"test"`
		err_expected	:= "Columns in CSV not equal"
		
		r := NewReader("")
		r.Col_integrity()
		_, err := r.Bytes([]byte(s), "")
		if err == nil {
			t.Fatal("Expected an error")
		}
		
		if err.Error() != err_expected {
			t.Fatalf("Expected error '%s', got '%v'", err_expected, err)
		}
		
		fmt.Println(strings.Join(r.Log(), "\n"))
	})
	
	t.Run("fill empty columns", func(t *testing.T){
		s 				:= `"test","test","test"`+"\n"+`"test"`+"\n"+`"test","test"`
		rows			:= "test,test,test\ntest,,\ntest,test,"
		
		r := NewReader("")
		r.Ignore_header()
		out, err := r.Bytes([]byte(s), "")
		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		
		var str []string
		for _, line := range out.Rows {
			str = append(str, strings.Join(line.Row, ","))
		}
		result := strings.Join(str, "\n")
		if result != rows {
			t.Fatalf("Want: %s\n\nGot: %s", rows, result)
		}
		
		fmt.Println(strings.Join(r.Log(), "\n"))
	})
	
	t.Run("column headers cannot be empty", func(t *testing.T){
		s 				:= `"head","","head"`+"\n"+`"test","test","test"`+"\n"+`"test","test","test"`
		err_expected	:= "Column headers cannot be empty"
		
		r := NewReader("")
		_, err := r.Bytes([]byte(s), "")
		if err == nil {
			t.Fatal("Expected an error")
		}
		
		if err.Error() != err_expected {
			t.Fatalf("Expected error '%s', got '%v'", err_expected, err)
		}
		
		fmt.Println(strings.Join(r.Log(), "\n"))
	})
	
	t.Run("column headers required", func(t *testing.T){
		s 				:= `"head","100,00","head"`+"\n"+`"test","test","test"`+"\n"+`"test","test","test"`
		err_expected	:= "Column headers in CSV required"
		
		r := NewReader("")
		_, err := r.Bytes([]byte(s), "")
		if err == nil {
			t.Fatal("Expected an error")
		}
		
		if err.Error() != err_expected {
			t.Fatalf("Expected error '%s', got '%v'", err_expected, err)
		}
		
		fmt.Println(strings.Join(r.Log(), "\n"))
	})
	
	t.Run("header and rows", func(t *testing.T){
		s 				:= `"head","head","head"`+"\n"+`"test","test","test"`+"\n"+`"test","test","test"`
		header 			:= "head,head,head"
		rows			:= "test,test,test\ntest,test,test"
		
		r := NewReader("")
		out, err := r.Bytes([]byte(s), "")
		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		
		result_header := strings.Join(out.Header, ",")
		if result_header != header {
			t.Fatalf("Want: %s\n\nGot: %s", header, result_header)
		}
		
		var str []string
		for _, line := range out.Rows {
			str = append(str, strings.Join(line.Row, ","))
		}
		result_rows := strings.Join(str, "\n")
		if result_rows != rows {
			t.Fatalf("Want: %s\n\nGot: %s", rows, result_rows)
		}
		
		fmt.Println(strings.Join(r.Log(), "\n"))
	})
	
	t.Run("too few column headers", func(t *testing.T){
		s 				:= `"head","head"`+"\n"+`"test","test","test"`+"\n"+`"test","test","test"`
		err_expected	:= "Too few column headers"
		
		r := NewReader("")
		_, err := r.Bytes([]byte(s), "")
		if err == nil {
			t.Fatal("Expected an error")
		}
		
		if err.Error() != err_expected {
			t.Fatalf("Expected error '%s', got '%v'", err_expected, err)
		}
		
		fmt.Println(strings.Join(r.Log(), "\n"))
	})
	
	t.Run("column headers cannot be empty", func(t *testing.T){
		s 				:= `"head","","head"`+"\n"+`"test","test","test"`+"\n"+`"test","test","test"`
		err_expected	:= "Column headers cannot be empty"
		
		r := NewReader("")
		_, err := r.Bytes([]byte(s), "")
		if err == nil {
			t.Fatal("Expected an error")
		}
		
		if err.Error() != err_expected {
			t.Fatalf("Expected error '%s', got '%v'", err_expected, err)
		}
		
		fmt.Println(strings.Join(r.Log(), "\n"))
	})
	
	t.Run("remove empty columns", func(t *testing.T){
		s 				:= `"head","head","head"`+"\n"+`"test","","test"`+"\n"+`"test","","test"`
		header 			:= "head,head"
		rows			:= "test,test\ntest,test"
		
		r := NewReader("")
		r.Remove_empty_cols()
		out, err := r.Bytes([]byte(s), "")
		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		
		result_header := strings.Join(out.Header, ",")
		if result_header != header {
			t.Fatalf("Want: %s\n\nGot: %s", header, result_header)
		}
		
		var str []string
		for _, line := range out.Rows {
			str = append(str, strings.Join(line.Row, ","))
		}
		result_rows := strings.Join(str, "\n")
		if result_rows != rows {
			t.Fatalf("Want: %s\n\nGot: %s", rows, result_rows)
		}
		
		fmt.Println(strings.Join(r.Log(), "\n"))
	})
}