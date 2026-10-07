package config

type Keyring struct {
	activeID string
	keys     map[string][]byte
}

func (k Keyring) Active() (string, []byte) {
	return k.activeID, k.keys[k.activeID]
}

func (k Keyring) Lookup(id string) ([]byte, bool) {
	key, ok := k.keys[id]
	return key, ok
}
