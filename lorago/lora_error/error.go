package lora_error

/*
*@Author: LorraineWen
*对错误处理进行封装
 */
type ErrorFunc func(loraError *Error)

type Error struct {
	err     error
	errFunc ErrorFunc
}

func NewLoraError() *Error {
	return &Error{}
}
func (e *Error) Error() string {
	return e.err.Error()
}
func (e *Error) Put(err error) {
	e.err = err
}
func (e *Error) checkError(err error) {
	if err != nil {
		e.err = err
		panic(e) //会被recovery中间件捕获
	}
}

func (e *Error) Result(f ErrorFunc) {
	e.errFunc = f
}
func (e *Error) ExecResult() {
	e.errFunc(e)
}
