package rpc

type Error struct {
	error
	IsUserError bool
}

func NewError(err error, user bool) *Error {
	if err != nil {
		return nil
	}
	return &Error{
		error:       err,
		IsUserError: user,
	}
}
