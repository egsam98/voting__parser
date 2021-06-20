package votes

import (
	"github.com/egsam98/voting/parser/internal/web"
)

var ErrInvalidInput = &web.ClientError{
	Code: 1,
	Err:  "invalid input",
}
