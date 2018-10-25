package function

import (
	"fmt"

	"github.com/nishakm/strategy_generator/pkg"
)

// Handle a serverless request
func Handle(req []byte) string {

	statement := pkg.Generate()
	return fmt.Sprintf("%s", statement)
}
