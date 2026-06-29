package model

import (
	"database/sql/driver"
	"fmt"
	"strings"

	"github.com/QuantumNous/new-api/common"
)

// StringList stores a JSON string array in a TEXT column while exposing a
// natural []string shape in API responses.
type StringList []string

func normalizeStringList(items []string) StringList {
	seen := make(map[string]struct{}, len(items))
	result := make(StringList, 0, len(items))
	for _, item := range items {
		normalized := strings.TrimSpace(item)
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	return result
}

func (l *StringList) Scan(value any) error {
	if value == nil {
		*l = StringList{}
		return nil
	}

	var raw string
	switch v := value.(type) {
	case string:
		raw = v
	case []byte:
		raw = string(v)
	default:
		return fmt.Errorf("unsupported StringList source %T", value)
	}

	raw = strings.TrimSpace(raw)
	if raw == "" {
		*l = StringList{}
		return nil
	}

	var items []string
	if err := common.Unmarshal([]byte(raw), &items); err != nil {
		// Legacy/manual fallback: tolerate comma-separated text.
		*l = normalizeStringList(strings.Split(raw, ","))
		return nil
	}
	*l = normalizeStringList(items)
	return nil
}

func (l StringList) Value() (driver.Value, error) {
	items := normalizeStringList([]string(l))
	data, err := common.Marshal([]string(items))
	if err != nil {
		return nil, err
	}
	return string(data), nil
}

func serializeStringList(items StringList) string {
	data, err := common.Marshal([]string(normalizeStringList([]string(items))))
	if err != nil {
		common.SysLog("failed to serialize string list: " + err.Error())
		return "[]"
	}
	return string(data)
}
