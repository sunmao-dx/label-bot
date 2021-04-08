package gitee_utils

type ErrorForbidden struct {
	err string
}

func (e ErrorForbidden) Error() string {
	return e.err
}
