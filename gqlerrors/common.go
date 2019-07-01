package gqlerrors

import (
	"strings"
)

type ForbiddenError struct {
	msg string
}

func (e *ForbiddenError) Error() string {
	if e.msg != "" {
		return e.msg
	}
	return "authentication failure, permission denied"
}

func (*ForbiddenError) Extensions() map[string]interface{} {
	return map[string]interface{}{
		"code": "FORBIDDEN",
	}
}

func Forbidden(msg string) error {
	return &ForbiddenError{msg}
}

type UnauthorizedError struct {
	msg string
}

func (e *UnauthorizedError) Error() string {
	if e.msg != "" {
		return e.msg
	}
	return "unauthorized user"
}

func (*UnauthorizedError) Extensions() map[string]interface{} {
	return map[string]interface{}{
		"code": "UNAUTHORIZED",
	}
}

func Unauthorized(msg string) error {
	return &UnauthorizedError{msg}
}

type BadArgumentError struct {
	arguments []string
}

func (e *BadArgumentError) Error() string {
	if len(e.arguments) > 0 {
		return "'" + strings.Join(e.arguments, "', '") + "' has some errors, please check it"
	}
	return "bad argument(s)"
}

func (*BadArgumentError) Extensions() map[string]interface{} {
	return map[string]interface{}{
		"code": "BAD_USER_INPUT",
	}
}

func UserInputError(arguments []string) error {
	return &BadArgumentError{arguments}
}

type InternalServerError struct {
	msg string
}

func (e *InternalServerError) Error() string {
	if e.msg != "" {
		return e.msg
	}
	return "internal server error"
}

func (*InternalServerError) Extensions() map[string]interface{} {
	return map[string]interface{}{
		"code": "INTERNAL_SERVER_ERROR",
	}
}

func InternalError(msg string) error {
	return &InternalServerError{msg}
}
