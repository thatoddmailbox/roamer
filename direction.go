package roamer

// A Direction describes the direction of a Migration or an Operation.
type Direction int

// The available directions.
const (
	DirectionUp   Direction = 1
	DirectionDown Direction = -1
)

// String returns a string representation of the Direction.
func (d Direction) String() string {
	switch d {
	case DirectionUp:
		return "up"
	case DirectionDown:
		return "down"
	}

	return "unknown"
}
