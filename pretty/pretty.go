package pretty

import (
	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
)

func Print(data any){
	b, _ := json.Marshal(data)
	(*jsontext.Value)(&b).Indent()
	fmt.Println(string(b))
}