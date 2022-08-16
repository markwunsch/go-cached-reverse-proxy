package cache

// TODO: implement local caching
type Local struct {
}

func NewLocal() *Local {
	return &Local{}
}

func (lc *Local) Put(key string, value string) error {
	return nil
}

func (lc *Local) Get(key string) (string, error) {
	return ``, nil
}
