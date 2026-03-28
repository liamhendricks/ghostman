package pipe_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/liamhendricks/ghostman/pkg/pipe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ReadJSON tests

func TestReadJSON_ValidInput(t *testing.T) {
	r := bytes.NewBufferString(`{"status":"ok"}`)
	data, err := pipe.ReadJSON(r)
	require.NoError(t, err)
	assert.Equal(t, `{"status":"ok"}`, string(data))
}

func TestReadJSON_EmptyInput(t *testing.T) {
	r := bytes.NewBufferString("")
	_, err := pipe.ReadJSON(r)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no input")
}

func TestReadJSON_InvalidJSON(t *testing.T) {
	r := bytes.NewBufferString("not json")
	_, err := pipe.ReadJSON(r)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid JSON")
}

// Extract tests

func TestExtract_NestedPath(t *testing.T) {
	data := []byte(`{"data":{"token":"abc123"}}`)
	val, err := pipe.Extract(data, "data.token")
	require.NoError(t, err)
	assert.Equal(t, "abc123", val)
}

func TestExtract_NestedPathWithLeadingDot(t *testing.T) {
	data := []byte(`{"data":{"token":"abc123"}}`)
	val, err := pipe.Extract(data, ".data.token")
	require.NoError(t, err)
	assert.Equal(t, "abc123", val)
}

func TestExtract_MissingPath(t *testing.T) {
	data := []byte(`{"data":{}}`)
	_, err := pipe.Extract(data, "data.missing")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "path not found")
}

func TestExtract_ArrayLength(t *testing.T) {
	data := []byte(`{"items":[1,2,3]}`)
	val, err := pipe.Extract(data, "items.#")
	require.NoError(t, err)
	assert.Equal(t, "3", val)
}

// Assert tests

func TestAssert_EqualMatch(t *testing.T) {
	data := []byte(`{"status":"ok"}`)
	err := pipe.Assert(data, `.status == "ok"`)
	assert.NoError(t, err)
}

func TestAssert_EqualMismatch(t *testing.T) {
	data := []byte(`{"status":"ok"}`)
	err := pipe.Assert(data, `.status == "fail"`)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `assertion failed: .status == "fail"`)
	assert.Contains(t, err.Error(), `got: "ok"`)
}

func TestAssert_NotEqualMatch(t *testing.T) {
	data := []byte(`{"status":"ok"}`)
	err := pipe.Assert(data, `.status != "fail"`)
	assert.NoError(t, err)
}

func TestAssert_NotEqualMismatch(t *testing.T) {
	data := []byte(`{"status":"ok"}`)
	err := pipe.Assert(data, `.status != "ok"`)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `assertion failed: .status != "ok"`)
}

func TestAssert_UnsupportedOperator(t *testing.T) {
	data := []byte(`{"status":"ok"}`)
	err := pipe.Assert(data, `.status > "ok"`)
	require.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "unsupported operator")
}

func TestAssert_MalformedExpression(t *testing.T) {
	data := []byte(`{"status":"ok"}`)
	err := pipe.Assert(data, "status")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid assertion")
}
