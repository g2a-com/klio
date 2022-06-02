package manager

import "fmt"

type CantFindExactVersionMatchError struct {
	depName, depVersion, depRegistry string
}

func (e *CantFindExactVersionMatchError) Error() string {
	return fmt.Sprintf("cannot find %s@%s in %s", e.depName, e.depVersion, e.depRegistry)
}
