package errors

type TaskError struct {
	SkipRetry bool
	Err       error
}

func (e TaskError) Error() string {
	return e.Cause().Error()
}

func (e *TaskError) Cause() error {
	return e.Err
}
