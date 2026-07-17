package imagegen

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

// ResolveAPIKey returns an explicit environment value or loads a key from a
// secret file. JSON fields use a dot-separated path such as
// dev.gemini_api_key. The credential is never included in an error.
func ResolveAPIKey(explicit, secretFile, field string) (string, error) {
	if strings.TrimSpace(explicit) != "" {
		return strings.TrimSpace(explicit), nil
	}
	if secretFile == "" {
		return "", errors.New("set GEMINI_API_KEY or configure a dev-secrets file")
	}
	data, err := os.ReadFile(secretFile)
	if err != nil {
		return "", fmt.Errorf("read imagegen secret file: %w", err)
	}
	if field == "" {
		key := strings.TrimSpace(string(data))
		if key == "" {
			return "", errors.New("imagegen secret file is empty")
		}
		return key, nil
	}
	var current any
	if err := json.Unmarshal(data, &current); err != nil {
		return "", errors.New("imagegen secret file is not valid JSON")
	}
	for _, segment := range strings.Split(field, ".") {
		object, ok := current.(map[string]any)
		if !ok {
			return "", fmt.Errorf("imagegen secret field %q does not exist", field)
		}
		current, ok = object[segment]
		if !ok {
			return "", fmt.Errorf("imagegen secret field %q does not exist", field)
		}
	}
	key, ok := current.(string)
	if !ok || strings.TrimSpace(key) == "" {
		return "", fmt.Errorf("imagegen secret field %q is not a non-empty string", field)
	}
	return strings.TrimSpace(key), nil
}
