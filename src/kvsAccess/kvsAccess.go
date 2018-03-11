package kvsAccess

func NewKVS() *KeyValStore {
    return &KeyValStore{Store: make(map[string]string)}
}

type KeyValStore struct {
    Store map[string]string
}

func (m *KeyValStore) NumKeys() (int) {
  return len(m.Store)
}

func (m *KeyValStore) SetValue(key string, newValue string) string {
    value, exists := m.Store[key]
    m.Store[key] = newValue
    if (exists) {
        return value
    } else {
        return ""
    }
}

func (m *KeyValStore) GetValue(key string) string {
    value, exists := m.Store[key]
    if (exists) {
        return value
    } else {
        return ""
    }
}

func (m KeyValStore) DelValue(key string) string {
    value, exists := m.Store[key]
    if (exists) {
        delete(m.Store, key)
        return value
    } else {
        return ""
    }
}
