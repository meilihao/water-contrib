package debug

import (
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/meilihao/logx"
	"github.com/meilihao/water"
)

func RequestDump(body bool) water.HandlerFunc {
	return func(ctx *water.Context) {
		// request Body is still ok after httputil.DumpRequest
		dump, err := httputil.DumpRequest(ctx.Request, body)
		if err != nil {
			http.Error(ctx, fmt.Sprint(err), http.StatusInternalServerError)
			return
		}

		logx.Debugf("requst in:\n%s\n",
			string(dump),
		)

		ctx.Next()
	}
}
