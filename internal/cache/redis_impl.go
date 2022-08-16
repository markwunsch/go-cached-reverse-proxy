package cache

type Redis struct {
}

// TODO: implement redis caching
func NewRedis() *Redis {
	return &Redis{}
}

func (rc *Redis) Put(key string, value string) error {
	return nil
}

func (rc *Redis) Get(key string) (string, error) {
	return ``, nil
}
