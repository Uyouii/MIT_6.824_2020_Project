package common

import "fmt"

const (
	Success int = iota
	UnknownTask
	InvalidReduceCnt
	InvalidWorkerId
	InvalidRequest
	InvalidTaskType
)

type Error struct {
	code int
	msg  string
}

func (e *Error) Error() string {
	msg := messages[e.code]
	return fmt.Sprintf("code: %v, msg: %v, %v", e.code, msg, e.msg)
}

func GetError(code int) error {
	return &Error{code: code}
}

func GetErrorWithMsg(code int, msg string) error {
	return &Error{code: code, msg: msg}
}

var messages = map[int]string{
	Success:          "success",
	UnknownTask:      "unknown task",
	InvalidReduceCnt: "invalid reduce task count",
	InvalidWorkerId:  "invalid worker id",
	InvalidRequest:   "invalid request",
	InvalidTaskType:  "invalid task type",
}
