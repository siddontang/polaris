package polaris

import (
	"fmt"
)

type HTTPError struct {
	Status  int    `json:"-"`
	Message string `json:"msg:`
	Result  string `json:"result"`
}

func (e HTTPError) Error() string {
	return fmt.Sprintf("Status: %d, Msg: %s, Result: %s", e.Status, e.Message, e.Result)
}
