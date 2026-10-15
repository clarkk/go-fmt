package sanitize

import "testing"

func Test_trim(t *testing.T){
	t.Run("disallow newlines", func(t *testing.T){
		input := 
`  some  text  
    
    
   width  a  newline  `
		
		want := "some text width a newline"
		
		got := Trim(input, false)
		
		if got != want {
			t.Fatalf("Want:\n%s\nGot:\n%s", want, got)
		}
	})
	
	t.Run("allow newlines", func(t *testing.T){
		input := 
`  some  text  
    
    
   width  a  newline  `
		
		want :=
`some text

width a newline`
		
		got := Trim(input, true)
		
		if got != want {
			t.Fatalf("Want:\n%s\nGot:\n%s", want, got)
		}
	})
	
	t.Run("allow newlines without", func(t *testing.T){
		input := "  some  text  "
		
		want := "some text"
		
		got := Trim(input, true)
		
		if got != want {
			t.Fatalf("Want:\n%s\nGot:\n%s", want, got)
		}
	})
	
	t.Run("disallow newlines without", func(t *testing.T){
		input := "  some  text  "
		
		want := "some text"
		
		got := Trim(input, false)
		
		if got != want {
			t.Fatalf("Want:\n%s\nGot:\n%s", want, got)
		}
	})
}

func Test_truncate(t *testing.T){
	t.Run("utf8", func(t *testing.T){
		input := "\xc3\xa6\xc3\xb8\xc3\xa5"
		
		want := "\xc3\xa6\xc3\xb8"
		
		got := Truncate(input, 2)
		
		if got != want {
			t.Fatalf("Want:\n%s\nGot:\n%s", want, got)
		}
	})
}