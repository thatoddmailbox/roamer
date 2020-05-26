package roamer

// A DriverType describes the type of database being used with roamer.
type DriverType string

// The available driver types.
const (
	DriverTypeMySQL DriverType = "mysql"
)

type driver interface {
	TableExists(name string) (bool, error)
}
