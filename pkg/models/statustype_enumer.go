// Code generated by "enumer -type=StatusType -json"; DO NOT EDIT.

//
package models

import (
	"encoding/json"
	"fmt"
)

const _StatusTypeName = "StatusInfoStatusSuccessStatusWarningStatusError"

var _StatusTypeIndex = [...]uint8{0, 10, 23, 36, 47}

func (i StatusType) String() string {
	if i < 0 || i >= StatusType(len(_StatusTypeIndex)-1) {
		return fmt.Sprintf("StatusType(%d)", i)
	}
	return _StatusTypeName[_StatusTypeIndex[i]:_StatusTypeIndex[i+1]]
}

var _StatusTypeValues = []StatusType{0, 1, 2, 3}

var _StatusTypeNameToValueMap = map[string]StatusType{
	_StatusTypeName[0:10]:  0,
	_StatusTypeName[10:23]: 1,
	_StatusTypeName[23:36]: 2,
	_StatusTypeName[36:47]: 3,
}

// StatusTypeString retrieves an enum value from the enum constants string name.
// Throws an error if the param is not part of the enum.
func StatusTypeString(s string) (StatusType, error) {
	if val, ok := _StatusTypeNameToValueMap[s]; ok {
		return val, nil
	}
	return 0, fmt.Errorf("%s does not belong to StatusType values", s)
}

// StatusTypeValues returns all values of the enum
func StatusTypeValues() []StatusType {
	return _StatusTypeValues
}

// IsAStatusType returns "true" if the value is listed in the enum definition. "false" otherwise
func (i StatusType) IsAStatusType() bool {
	for _, v := range _StatusTypeValues {
		if i == v {
			return true
		}
	}
	return false
}

// MarshalJSON implements the json.Marshaler interface for StatusType
func (i StatusType) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.String())
}

// UnmarshalJSON implements the json.Unmarshaler interface for StatusType
func (i *StatusType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("StatusType should be a string, got %s", data)
	}

	var err error
	*i, err = StatusTypeString(s)
	return err
}
