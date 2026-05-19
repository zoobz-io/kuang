package secrets

import "github.com/zoobz-io/check"

// Secret is the response for retrieving a secret by key.
type Secret struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Validate validates a Secret.
func (s Secret) Validate() error {
	return check.All(
		check.Str(s.Key, "key").Required().V(),
		check.Str(s.Value, "value").Required().V(),
	).Err()
}
