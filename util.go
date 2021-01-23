package cmdsrv

import (
	"reflect"
	"runtime"
)

func funcName(f interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}
