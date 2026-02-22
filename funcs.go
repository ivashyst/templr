package templr

import (
	"errors"
	"html/template"
	"time"
)

var funcMap = template.FuncMap{
	"formatDate": func(t time.Time) string {
		return t.Format("2006-01-02")
	},
	"formatDateTime": func(t time.Time) string {
		return t.Format("2006-01-02 15:04")
	},
	"formatTime": func(t time.Time) string {
		return t.Format("15:04")
	},

	"dict": func(values ...any) (map[string]any, error) {
		if len(values)%2 != 0 {
			return nil, errors.New("invalid dict")
		}

		result := make(map[string]any)
		for i := 0; i < len(values); i += 2 {
			key, ok := values[i].(string)
			if !ok {
				return nil, errors.New("keys must be strings")
			}
			result[key] = values[i+1]
		}
		return result, nil
	},
}
