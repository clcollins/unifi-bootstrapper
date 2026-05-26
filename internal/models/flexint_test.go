package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlexInt_UnmarshalFromInt(t *testing.T) {
	input := []byte(`123`)
	var fi FlexInt
	err := json.Unmarshal(input, &fi)
	require.NoError(t, err)
	assert.Equal(t, 123, fi.Int())
}

func TestFlexInt_UnmarshalFromString(t *testing.T) {
	input := []byte(`"456"`)
	var fi FlexInt
	err := json.Unmarshal(input, &fi)
	require.NoError(t, err)
	assert.Equal(t, 456, fi.Int())
}

func TestFlexInt_UnmarshalFromZeroInt(t *testing.T) {
	input := []byte(`0`)
	var fi FlexInt
	err := json.Unmarshal(input, &fi)
	require.NoError(t, err)
	assert.Equal(t, 0, fi.Int())
}

func TestFlexInt_UnmarshalFromZeroString(t *testing.T) {
	input := []byte(`"0"`)
	var fi FlexInt
	err := json.Unmarshal(input, &fi)
	require.NoError(t, err)
	assert.Equal(t, 0, fi.Int())
}

func TestFlexInt_UnmarshalInvalidString(t *testing.T) {
	input := []byte(`"not-a-number"`)
	var fi FlexInt
	err := json.Unmarshal(input, &fi)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot parse string")
}

func TestFlexInt_UnmarshalInvalidJSON(t *testing.T) {
	input := []byte(`[1,2,3]`)
	var fi FlexInt
	err := json.Unmarshal(input, &fi)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot unmarshal")
}

func TestFlexInt_MarshalAlwaysInt(t *testing.T) {
	fi := FlexInt(789)
	data, err := json.Marshal(fi)
	require.NoError(t, err)
	assert.Equal(t, `789`, string(data))
}

func TestFlexInt_RoundTrip(t *testing.T) {
	original := FlexInt(2000)
	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded FlexInt
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, original, decoded)
}

func TestFlexInt_InStruct(t *testing.T) {
	type wrapper struct {
		Index FlexInt `json:"index"`
	}

	tests := []struct {
		name  string
		input string
		want  int
	}{
		{
			name:  "integer value in struct",
			input: `{"index":2001}`,
			want:  2001,
		},
		{
			name:  "string value in struct",
			input: `{"index":"3000"}`,
			want:  3000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var w wrapper
			err := json.Unmarshal([]byte(tt.input), &w)
			require.NoError(t, err)
			assert.Equal(t, tt.want, w.Index.Int())
		})
	}
}
