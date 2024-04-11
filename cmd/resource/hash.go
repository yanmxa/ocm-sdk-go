package resource

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

func StatusHashGetter(obj *Resource) (string, error) {
	statusBytes, err := json.Marshal(obj.Status.Conditions)
	if err != nil {
		return "", fmt.Errorf("failed to marshal resource status, %v", err)
	}
	return fmt.Sprintf("%x", sha256.Sum256(statusBytes)), nil
}
