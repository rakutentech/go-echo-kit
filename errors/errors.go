package errors

import (
	"errors"
	"fmt"
)

// ErrorCode - refer to const for detail
type ErrorCode string

// Error - contains build-in error and ErrorCode
type Error struct {
	error     error
	errorCode ErrorCode
}

// NewError - creates an Error instance with built-in error
func NewError(code ErrorCode, err error) Error {
	return Error{error: err, errorCode: code}
}

// NewErrorWithMsg - creates an Error instance with custom message
func NewErrorWithMsg(code ErrorCode, msg string) Error {
	return NewError(code, errors.New(msg))
}

// NewErrorWithMsgf - creates an Error instance with formatted message
func NewErrorWithMsgf(code ErrorCode, format string, values ...interface{}) Error {
	msg := fmt.Sprintf(format, values...)
	return NewError(code, errors.New(msg))
}

// ErrorCode - get the ErrorCode of an Error instance
// Usage err.(errors.Error).ErrorCode()
func (err Error) ErrorCode() string {
	return string(err.errorCode)
}

// Implementation of built-in error interface
func (err Error) Error() string {
	return err.error.Error()
}

// Error Codes
const (
	ErrCodeParameterIllegalState   ErrorCode = "IllegalArgument"
	ErrCodeSQLResult               ErrorCode = "SQL_Result"
	ErrCodeSQLIllegalState         ErrorCode = "SQL_IllegalState"
	ErrCodeEmailIllegalState       ErrorCode = "Email_IllegalState"
	ErrCodeMonitoringAbnormalState ErrorCode = "Monitoring_AbnormalState"

	ErrCodeMissingParamsError   ErrorCode = "MissingParams_Error"
	ErrCodeValidationError      ErrorCode = "Validation_Error"
	ErrCodeDuplicateParamsError ErrorCode = "DuplicateParams_Error"
	ErrCodeUnexpectedError      ErrorCode = "Unexpected_Error"

	ErrCodeHTTPRequestIllegalState ErrorCode = "HTTPRequest_IllegalState"
	ErrCodeFileAlreadyExists       ErrorCode = "FileAlreadyExists_Error"
	ErrCodeAlreadyExistsInDB       ErrorCode = "AlreadyExistsInDB_Error"
	ErrCodeNotExistInDB            ErrorCode = "NotExistInDB_Error"
)
