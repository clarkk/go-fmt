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
}