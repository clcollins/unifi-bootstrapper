package models

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// FlexInt handles JSON values that may be encoded as either a string
// or an integer. The UDM-Pro API returns rule_index as a string on
// some firmware versions and as an integer on others. FlexInt
// normalizes both representations to an int.
type FlexInt int

// UnmarshalJSON implements json.Unmarshaler for FlexInt. It accepts
// both JSON strings (e.g., "2000") and JSON numbers (e.g., 2000).
func (fi *FlexInt) UnmarshalJSON(data []byte) error {
	// Try unmarshaling as int first (most common)
	var intVal int
	if err := json.Unmarshal(data, &intVal); err == nil {
		*fi = FlexInt(intVal)
		return nil
	}

	// Try unmarshaling as string
	var strVal string
	if err := json.Unmarshal(data, &strVal); err == nil {
		parsed, err := strconv.Atoi(strVal)
		if err != nil {
			return fmt.Errorf("FlexInt: cannot parse string %q as int: %w", strVal, err)
		}
		*fi = FlexInt(parsed)
		return nil
	}

	return fmt.Errorf("FlexInt: cannot unmarshal %s", string(data))
}

// MarshalJSON implements json.Marshaler for FlexInt. It always
// marshals as a JSON integer.
func (fi FlexInt) MarshalJSON() ([]byte, error) {
	return json.Marshal(int(fi))
}

// Int returns the FlexInt value as a plain int.
func (fi FlexInt) Int() int {
	return int(fi)
}
