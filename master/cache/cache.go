package cache

type Cache interface {
	Get(key string) (string, bool)
	Put(key string, value string)
	Delete(key string)
	Size() int
}
