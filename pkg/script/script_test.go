package script_test

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"testing"

	"github.com/liamhendricks/ghostman/pkg/script"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHelperProcess is NOT a real test. It is a helper process used by the
// exec-self tests below. It runs as a subprocess.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	switch os.Getenv("HELPER_CASE") {
	case "run_success":
		script.Run(setFieldHandler{"processed", "true"})
	case "run_invalid_json":
		// stdin will be "not json"
		script.Run(setFieldHandler{"processed", "true"})
	case "run_handler_error":
		script.RunFunc(func(d script.Data) (script.Data, error) {
			return d, assert.AnError
		})
	case "runfunc_success":
		script.RunFunc(func(d script.Data) (script.Data, error) {
			return d.Set("ok", "yes")
		})
	}
	os.Exit(0)
}

// setFieldHandler is a Handler used in helper processes.
type setFieldHandler struct {
	field string
	value string
}

func (h setFieldHandler) Handle(d script.Data) (script.Data, error) {
	return d.Set(h.field, h.value)
}

// ---- Data unit tests ----

func TestNewData(t *testing.T) {
	raw := []byte(`{"a":"b"}`)
	d := script.NewData(raw)
	assert.Equal(t, raw, d.Raw())
}

func TestData_Get(t *testing.T) {
	d := script.NewData([]byte(`{"data":{"token":"abc"}}`))
	assert.Equal(t, "abc", d.Get("data.token"))
}

func TestData_Get_TopLevel(t *testing.T) {
	d := script.NewData([]byte(`{"name":"test"}`))
	assert.Equal(t, "test", d.Get("name"))
}

func TestData_Get_Missing(t *testing.T) {
	d := script.NewData([]byte(`{"a":"b"}`))
	assert.Equal(t, "", d.Get("missing"))
}

func TestData_Set(t *testing.T) {
	original := script.NewData([]byte(`{"name":"test"}`))
	updated, err := original.Set("age", "25")
	require.NoError(t, err)
	assert.Equal(t, "25", updated.Get("age"))
	// Original is unchanged
	assert.Equal(t, "", original.Get("age"))
}

func TestData_Set_Nested(t *testing.T) {
	d := script.NewData([]byte(`{}`))
	updated, err := d.Set("a.b", "val")
	require.NoError(t, err)
	assert.Equal(t, "val", updated.Get("a.b"))
}

func TestData_Raw(t *testing.T) {
	raw := []byte(`{"x":1}`)
	d := script.NewData(raw)
	assert.Equal(t, raw, d.Raw())
}

// ---- Run/RunFunc exec-self tests ----

func runHelperProcess(t *testing.T, helperCase string, stdin string) ([]byte, int, error) {
	t.Helper()
	// -test.v=false suppresses test framework output so only the handler's
	// stdout (the JSON result) is captured by cmd.Output().
	cmd := exec.Command(os.Args[0], "-test.run=TestHelperProcess", "-test.v=false")
	cmd.Env = append(os.Environ(),
		"GO_WANT_HELPER_PROCESS=1",
		"HELPER_CASE="+helperCase,
	)
	cmd.Stdin = bytes.NewBufferString(stdin)
	out, err := cmd.Output()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	}
	return out, exitCode, err
}

func TestRun_Success(t *testing.T) {
	out, exitCode, err := runHelperProcess(t, "run_success", `{"in":1}`)
	require.NoError(t, err)
	assert.Equal(t, 0, exitCode)
	assert.True(t, json.Valid(out), "output should be valid JSON")
	d := script.NewData(out)
	assert.Equal(t, "true", d.Get("processed"))
}

func TestRun_InvalidJSON(t *testing.T) {
	_, exitCode, _ := runHelperProcess(t, "run_invalid_json", "not json")
	assert.NotEqual(t, 0, exitCode, "should exit non-zero on invalid stdin JSON")
}

func TestRun_HandlerError(t *testing.T) {
	_, exitCode, _ := runHelperProcess(t, "run_handler_error", `{"x":1}`)
	assert.NotEqual(t, 0, exitCode, "should exit non-zero when handler returns error")
}

func TestRunFunc_Success(t *testing.T) {
	out, exitCode, err := runHelperProcess(t, "runfunc_success", `{"x":1}`)
	require.NoError(t, err)
	assert.Equal(t, 0, exitCode)
	assert.True(t, json.Valid(out), "output should be valid JSON")
	d := script.NewData(out)
	assert.Equal(t, "yes", d.Get("ok"))
}
