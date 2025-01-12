package storage

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type Timestamp struct {
	val time.Time
}

func NewTimestamp(val time.Time) Timestamp {
	return Timestamp{
		val: val.Truncate(time.Microsecond),
	}
}

func NewTimestampPtr(val time.Time) *Timestamp {
	return &Timestamp{
		val: val.Truncate(time.Microsecond),
	}
}

func (t *Timestamp) ToTime() time.Time {
	if t == nil {
		return time.Time{}
	}

	return t.val
}

func (t *Timestamp) Scan(src interface{}) error {
	if src == nil {
		return nil
	}
	val, ok := src.(time.Time)
	if !ok {
		return fmt.Errorf("failed to unmarshal timestamp value: %s", src)
	}
	t.val = val
	return nil
}

func (t Timestamp) Value() (driver.Value, error) {
	return t.val, nil
}

type JSONB struct {
	val []byte
}

func NewJSONB(val any) (*JSONB, error) {
	if val == nil {
		return nil, nil
	}

	jsonb := new(JSONB)
	switch v := val.(type) {
	case []byte:
		valid := json.Valid(v)
		if !valid {
			return nil, fmt.Errorf("invalid json value: %s", v)
		}

		jsonb.val = v
	case string:
		bytes := []byte(v)
		valid := json.Valid(bytes)
		if !valid {
			return nil, fmt.Errorf("invalid json value: %s", bytes)
		}

		jsonb.val = bytes
	case *string:
		if v != nil {
			bytes := []byte(*v)
			valid := json.Valid(bytes)
			if !valid {
				return nil, fmt.Errorf("invalid json value: %s", bytes)
			}

			jsonb.val = bytes
		}
	default:
		marshal, err := json.Marshal(val)
		if err != nil {
			return nil, fmt.Errorf("marshal: %w", err)
		}
		jsonb.val = marshal
	}

	return jsonb, nil
}

func (v *JSONB) Byte() []byte {
	if v == nil {
		return nil
	}
	return v.val
}

func (v *JSONB) Scan(src interface{}) error {
	if src == nil {
		return nil
	}
	switch val := src.(type) {
	case []byte:
		v.val = val
	case string:
		v.val = []byte(val)
	default:
		return fmt.Errorf("failed to unmarshal JSONB value: %v", src)
	}

	return nil
}

func (v JSONB) Value() (driver.Value, error) {
	return string(v.val), nil
}
