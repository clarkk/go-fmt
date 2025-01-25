package csv

type Error struct {
	s 	string
	err error
}

func (e *Error) Error() string {
	return e.s
}

func (e *Error) Unwrap() error {
	return e.err
}