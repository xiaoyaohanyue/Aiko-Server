package mydispatcher

import "github.com/AikoPanel/Xray-core/common/errors"

func newError(values ...interface{}) *errors.Error {
	return errors.New(values...)
}
