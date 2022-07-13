package z

import "github.com/snwfdhmp/errlog"

func Err(err error) bool {
	return errlog.Debug(err)
}
