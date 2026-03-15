package shared

import "fmt"

type Button int

const (
	ButtonLeft Button = iota
	ButtonMiddle
	ButtonRight
)

var buttonNames = [...]string{
	"left",
	"middle",
	"right",
}

func (b Button) String() string {
	if b.Valid() {
		return buttonNames[b]
	}
	return fmt.Sprintf("Button(%d)", int(b))
}

func (b Button) Valid() bool {
	return b >= 0 && int(b) < len(buttonNames)
}
