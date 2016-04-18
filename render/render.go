package render

import (
	"github.com/meilihao/water"
)

type Render interface {
	HTML(*water.Context, string, map[string]interface{}) error
}
