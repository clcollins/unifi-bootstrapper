package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPIResponse_UnmarshalSuccess(t *testing.T) {
	input := `{
		"meta": {"rc": "ok"},
		"data": [{"name": "test"}]
	}`

	type simple struct {
		Name string `json:"name"`
	}

	var resp APIResponse[simple]
	err := json.Unmarshal([]byte(input), &resp)
	require.NoError(t, err)
	assert.Equal(t, "ok", resp.Meta.RC)
	require.Len(t, resp.Data, 1)
	assert.Equal(t, "test", resp.Data[0].Name)
}

func TestAPIResponse_UnmarshalEmptyData(t *testing.T) {
	input := `{
		"meta": {"rc": "ok"},
		"data": []
	}`

	type simple struct {
		Name string `json:"name"`
	}

	var resp APIResponse[simple]
	err := json.Unmarshal([]byte(input), &resp)
	require.NoError(t, err)
	assert.Equal(t, "ok", resp.Meta.RC)
	assert.Empty(t, resp.Data)
}

func TestAPIResponse_UnmarshalError(t *testing.T) {
	input := `{
		"meta": {"rc": "error"},
		"data": []
	}`

	type simple struct {
		Name string `json:"name"`
	}

	var resp APIResponse[simple]
	err := json.Unmarshal([]byte(input), &resp)
	require.NoError(t, err)
	assert.Equal(t, "error", resp.Meta.RC)
}

func TestAPIResponse_RoundTrip(t *testing.T) {
	type simple struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	original := APIResponse[simple]{
		Data: []simple{
			{Name: "alpha", Value: 1},
			{Name: "beta", Value: 2},
		},
	}
	original.Meta.RC = "ok"

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded APIResponse[simple]
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, original, decoded)
}
