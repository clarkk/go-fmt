package pretty

import (
	"fmt"
	"encoding/json/v2"
	"encoding/json/jsontext"
)

func Print(data any){
	fmt.Println(String(data))
}

func String(data any) string {
	b, _ := json.Marshal(data)
	(*jsontext.Value)(&b).Indent()
	return string(b)
}