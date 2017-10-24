package function

import (
	"fmt"

	"github.com/mlabouardy/9gag"
)

func Handle(req []byte) string {
	gag9 := gag9.New()
	memes := gag9.FindByTag(string(req))
	return fmt.Sprintf("%s", memes)
}
