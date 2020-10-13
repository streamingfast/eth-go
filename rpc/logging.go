package rpc

import (
	"os"

	"github.com/dfuse-io/logging"
	"go.uber.org/zap"
)

var traceEnabled = os.Getenv("TRACE") == "true"
var zlog = zap.NewNop()

func init() {
	logging.Register("github.com/dfuse-io/eth-go/rpc", &zlog)
}
