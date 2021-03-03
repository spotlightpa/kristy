package healthcheckio

import (
	"fmt"
	"net/http"
)

// StatusErr represents an unexpected response status from HealthCheck.io
type StatusErr int

func (se StatusErr) String() string {
	return http.StatusText(int(se))
}

func (se StatusErr) Error() string {
	return fmt.Sprintf("unexpected status: %d %s",
		int(se), se.String())
}
