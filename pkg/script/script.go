package script

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// Data holds raw JSON bytes and provides path-based access via gjson/sjson.
type Data struct {
	raw []byte
}

// NewData creates a new Data from raw JSON bytes.
func NewData(raw []byte) Data {
	return Data{raw: raw}
}

// Raw returns the underlying JSON bytes.
func (d Data) Raw() []byte {
	return d.raw
}

// Get extracts the string value at path using gjson dot-notation.
// Returns an empty string if the path does not exist.
func (d Data) Get(path string) string {
	return gjson.GetBytes(d.raw, path).String()
}

// Set returns a new Data with the value at path set to value using sjson.
// The original Data is not modified.
func (d Data) Set(path, value string) (Data, error) {
	updated, err := sjson.SetBytes(d.raw, path, value)
	if err != nil {
		return d, err
	}
	return Data{raw: updated}, nil
}

// Handler is the interface that user scripts implement.
type Handler interface {
	Handle(data Data) (Data, error)
}

// Run reads JSON from stdin, calls h.Handle with the data, and writes the
// result to stdout. Exits non-zero on invalid stdin JSON or handler error.
func Run(h Handler) {
	raw, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "script: read stdin: %v\n", err)
		os.Exit(1)
	}
	if !json.Valid(raw) {
		fmt.Fprintln(os.Stderr, "script: invalid JSON on stdin")
		os.Exit(1)
	}
	result, err := h.Handle(Data{raw: raw})
	if err != nil {
		fmt.Fprintf(os.Stderr, "script: %v\n", err)
		os.Exit(1)
	}
	os.Stdout.Write(result.Raw()) //nolint:errcheck
}

// RunFunc wraps a function literal as a Handler and delegates to Run.
func RunFunc(fn func(Data) (Data, error)) {
	Run(handlerFunc(fn))
}

// handlerFunc adapts a function to the Handler interface.
type handlerFunc func(Data) (Data, error)

func (f handlerFunc) Handle(d Data) (Data, error) {
	return f(d)
}
