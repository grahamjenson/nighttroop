package nightscout

import "encoding/json"

func FromMap[T any](m map[string]any, dst *T) error {
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, dst)
}

func ToMap[T any](src *T) (map[string]any, error) {
	b, err := json.Marshal(src)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	err = json.Unmarshal(b, &m)
	return m, err
}
