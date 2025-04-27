package botkit

import (
	"encoding/json"
	"strings"
)

func ParseJSON[T any](src string) (T, error) {
	var args T

	// удаление переносов строк в телеграм
	cleaned := strings.ReplaceAll(src, "\n", "")
	cleaned = strings.ReplaceAll(cleaned, "\r", "")

	if err := json.Unmarshal([]byte(cleaned), &args); err != nil {
		return *(new(T)), err
	}

	return args, nil
}
