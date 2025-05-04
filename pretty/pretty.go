package pretty

import (
	"fmt"
	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
)

func Print(data any){
	fmt.Println(String(data))
}

func String(data any) string {
	b, _ := json.Marshal(data)
	(*jsontext.Value)(&b).Indent()
	return string(b)
}