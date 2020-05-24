package roamer

type driver interface {
	TableExists(name string) (bool, error)
}
