package healthcheckio

import (
	"fmt"
	"net/http"
)

// StatusErr is an unexpected status
type StatusErr int

func (se StatusErr) String() string {
	return http.StatusText(int(se))
}

func (se StatusErr) Error() string {
	return fmt.Sprintf("unexpected status: %d %s",
		int(se), se.String())
}
