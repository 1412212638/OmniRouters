package jsplugin

import (
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/relaykit/relayconvert/kitutil"
	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type bytePreservingJSONCodec struct {
	jsoniter.API
}

func (c bytePreservingJSONCodec) Decode(reader io.Reader, value any) error {
	return c.NewDecoder(reader).Decode(value)
}

func TestJSONStateUTF8Normalization(t *testing.T) {
	const codecEnv = "JSPLUGIN_TEST_BYTE_PRESERVING_JSON"
	if os.Getenv(codecEnv) == "1" {
		// Only this test runs in the child process; the parent's codec and
		// parallel tests are unaffected. Disabling HTML escaping in jsoniter
		// also preserves invalid UTF-8, independently of Sonic's platform support.
		kitutil.SetCodec(bytePreservingJSONCodec{jsoniter.Config{SortMapKeys: true}.Froze()})
		encoded, err := common.Marshal("\xff\xfe")
		require.NoError(t, err)
		require.Equal(t, []byte("\"\xff\xfe\""), encoded)
		var decoded any
		require.NoError(t, common.Unmarshal(encoded, &decoded))
		require.Equal(t, "\xff\xfe", decoded)
	} else {
		t.Run("byte preserving codec", func(t *testing.T) {
			executable, err := os.Executable()
			require.NoError(t, err)
			command := exec.CommandContext(t.Context(), executable, "-test.run=^TestJSONStateUTF8Normalization$", "-test.v")
			command.Env = append(os.Environ(), codecEnv+"=1")
			output, err := command.CombinedOutput()
			require.NoError(t, err, "%s", output)
		})
	}

	const raw = "图像\xff\xfe😀<&\n\"\\\u2028"
	const normalized = "图像\ufffd\ufffd😀<&\n\"\\\u2028"
	for _, tc := range []struct {
		name    string
		op      string
		path    []any
		initial any
		value   any
		want    any
	}{
		{"set", "set", []any{}, false, raw, normalized},
		{"append", "append", []any{}, []any{float64(0), false}, raw, []any{float64(0), false, normalized}},
		{"appendText", "appendText", []any{"a<b"}, map[string]any{"a<b": "前\xff"}, raw, map[string]any{"a<b": "前\ufffd" + normalized}},
		{"separate invalid bytes", "appendText", []any{}, "\xff", "\xfe", "\ufffd\ufffd"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			encoded, err := common.Marshal(tc.want)
			require.NoError(t, err)
			for _, shortBy := range []int{0, 1} {
				state := NewJSONState(len(encoded) - shortBy)
				err := state.Apply(t.Context(), []any{
					map[string]any{"op": "set", "path": []any{}, "value": tc.initial},
					map[string]any{"op": tc.op, "path": tc.path, "value": tc.value},
				})
				if shortBy == 0 {
					require.NoError(t, err, "normalized JSON must fit exactly")
					value, err := state.Value()
					require.NoError(t, err)
					assert.Equal(t, tc.want, value)
					continue
				}
				require.Error(t, err, "normalized JSON must exceed the limit by one byte")
				value, valueErr := state.Value()
				require.ErrorIs(t, valueErr, err)
				assert.Nil(t, value, "a failed batch must not publish its successful prefix")
				require.ErrorIs(t, state.Apply(t.Context(), []any{
					map[string]any{"op": "set", "path": []any{}, "value": nil},
				}), err, "a later replacement must not revive failed state")
			}
		})
	}

	t.Run("clone", func(t *testing.T) {
		engine, err := Compile(`export function clone(input) { return utils.json.clone(input); }`, Options{})
		require.NoError(t, err)
		input := map[string]any{"text": raw, "items": []any{float64(0), false, raw}}
		value, err := engine.Call(t.Context(), "clone", input)
		require.NoError(t, err)
		assert.Equal(t, map[string]any{"text": normalized, "items": []any{int64(0), false, normalized}}, value)
		assert.Equal(t, raw, input["text"])

		encoded, err := common.Marshal(normalized)
		require.NoError(t, err)
		padding := strings.Repeat("x", MaxJSONToolBytes-len(encoded))
		value, err = engine.Call(t.Context(), "clone", raw+padding)
		require.NoError(t, err, "normalized clone must fit exactly")
		assert.Equal(t, normalized, strings.TrimSuffix(value.(string), padding))
		value, err = engine.Call(t.Context(), "clone", raw+padding+"x")
		require.Error(t, err, "normalized clone must exceed the limit by one byte")
		assert.Nil(t, value)
	})

	t.Run("invalid keys and paths remain rejected", func(t *testing.T) {
		for _, change := range []any{
			map[string]any{"op": "set", "path": []any{}, "value": map[string]any{"\xff\xfe": raw}},
			map[string]any{"op": "set", "path": []any{"\xff\xfe"}, "value": raw},
		} {
			state := NewJSONState(128)
			require.Error(t, state.Apply(t.Context(), []any{
				map[string]any{"op": "set", "path": []any{}, "value": map[string]any{"\ufffd\ufffd": "original"}},
				change,
			}))
			value, err := state.Value()
			require.Error(t, err)
			assert.Nil(t, value)
		}
	})
}
