package cache

// Manager is an interface to abstract the cache layer
type Manager interface {
	Put(key string, value string) error
	Get(key string) (string, error)
}
