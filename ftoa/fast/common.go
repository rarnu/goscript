package fast

import "errors"

var (
	dcheckFailure = errors.New("DCHECK assertion failed")
)

func _DCHECK(f bool) {
	if !f {
		panic(dcheckFailure)
	}
}
