// example.go demonstrates a ghostman script using the RunFunc helper.
//
// Usage:
//
//	echo '{"name":"world"}' | ghostman script .ghostman/scripts/example.go
//
// Input:  {"name":"world"}
// Output: {"name":"world","greeting":"Hello, world!"}
package main

import "github.com/liamhendricks/ghostman/pkg/script"

func main() {
	script.RunFunc(func(d script.Data) (script.Data, error) {
		name := d.Get("name")
		println("example.go script RunFunc")
		return d.Set("greeting", "Hello, "+name+"!")
	})
}
