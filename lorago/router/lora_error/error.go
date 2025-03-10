package lora_error

/*
*@Author: LorraineWen
*对错误处理进行封装
 */
type ErrorFunc func(loraError *LoraError)

type LoraError struct {
	err     error
	errFunc ErrorFunc
}

func NewLoraError() *LoraError {
	return &LoraError{}
}
func (e *LoraError) Error() string {
	return e.err.Error()
}
func (e *LoraError) Put(err error) {
	e.err = err
}
func (e *LoraError) checkError(err error) {
	if err != nil {
		e.err = err
		panic(e) //会被recovery中间件捕获
	}
}

func (e *LoraError) Result(f ErrorFunc) {
	e.errFunc = f
}
func (e *LoraError) ExecResult() {
	e.errFunc(e)
}
