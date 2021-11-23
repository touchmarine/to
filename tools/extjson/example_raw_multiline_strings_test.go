package extjson_test

import (
	"fmt"
	"strings"

	"github.com/touchmarine/to/tools/extjson"
)

const src = `
"Templates": {
	"html": '''
<blockquote>
	{{template "children" .}}
</blockquote>
	'''
}
`

func Example() {
	var b strings.Builder
	extjson.Convert(&b, strings.NewReader(src))
	fmt.Println(b.String())

	// Output:
	// "Templates": {
	// 	"html": "<blockquote>\n\t{{template \"children\" .}}\n</blockquote>\n\t"
	// }
}
