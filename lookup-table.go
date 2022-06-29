package echo_binder

import (
	"encoding/json"
	"strings"
)

type lookupTable map[string]interface{}

type RecursiveLookupTable map[string]RecursiveLookupTable

func (l *lookupTable) FieldExists(key string) bool {
	keys := strings.Split(key, ".")
	if len(keys) == 0 {
		return false
	}

	data, ok := (*l)[keys[0]]
	if len(keys) == 1 || !ok {
		return ok
	}

	switch t := data.(type) {
	case lookupTable:
		return l.FieldExists(strings.Join(keys[1:], "."))

	default:
		data, err := json.Marshal(&t)
		if err != nil {
			return false
		}

		lut := lookupTable{}
		if err = json.Unmarshal(data, &lut); err != nil {
			return false
		}

		return lut.FieldExists(strings.Join(keys[1:], "."))
	}
}

func (l *lookupTable) IntoRecursiveLookupTable() RecursiveLookupTable {
	rlt := RecursiveLookupTable{}

	for key, value := range *l {
		switch v := value.(type) {
		case lookupTable:
			rlt[key] = v.IntoRecursiveLookupTable()

		default:
			data, err := json.Marshal(&v)
			if err != nil {
				rlt[key] = RecursiveLookupTable{}
				continue
			}

			lut := lookupTable{}
			if err = json.Unmarshal(data, &lut); err != nil {
				rlt[key] = RecursiveLookupTable{}
				continue
			}

			rlt[key] = lut.IntoRecursiveLookupTable()
		}
	}

	return rlt
}

func (l *RecursiveLookupTable) FieldExists(key string) bool {
	keys := strings.Split(key, ".")
	if len(keys) == 0 {
		return false
	}

	data, ok := (*l)[keys[0]]
	if len(keys) == 1 || !ok {
		return ok
	}

	return data.FieldExists(strings.Join(keys[1:], "."))
}
