package cli

import (
	"encoding/json"
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

const customTimeDefaultFormat = "Jan 2"

type CustomTime struct {
	time.Time
}

func (t CustomTime) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%q", t.Format(customTimeDefaultFormat))), nil
}
func (t *CustomTime) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	ts, err := time.Parse(customTimeDefaultFormat, s)
	if err != nil {
		return err
	}

	*t = CustomTime{ts}
	return nil
}

func (t CustomTime) MarshalYAML() (interface{}, error) {
	return t.Format(customTimeDefaultFormat), nil
}

func (t *CustomTime) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return err
	}

	ts, err := time.Parse(customTimeDefaultFormat, s)
	if err != nil {
		return err
	}
	*t = CustomTime{ts}
	return nil
}
