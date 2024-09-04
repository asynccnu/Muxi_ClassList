// Code generated by protoc-gen-go-errors. DO NOT EDIT.

package v1

import (
	fmt "fmt"
	errors "github.com/go-kratos/kratos/v2/errors"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the kratos package it is being compiled against.
const _ = errors.SupportPackageIsVersion1

func IsDbNotfound(err error) bool {
	if err == nil {
		return false
	}
	e := errors.FromError(err)
	return e.Reason == ErrorReason_DB_NOTFOUND.String() && e.Code == 200
}

func ErrorDbNotfound(format string, args ...interface{}) *errors.Error {
	return errors.New(200, ErrorReason_DB_NOTFOUND.String(), fmt.Sprintf(format, args...))
}

func IsDbFinderr(err error) bool {
	if err == nil {
		return false
	}
	e := errors.FromError(err)
	return e.Reason == ErrorReason_DB_FINDERR.String() && e.Code == 450
}

func ErrorDbFinderr(format string, args ...interface{}) *errors.Error {
	return errors.New(450, ErrorReason_DB_FINDERR.String(), fmt.Sprintf(format, args...))
}

func IsDbUpdateerr(err error) bool {
	if err == nil {
		return false
	}
	e := errors.FromError(err)
	return e.Reason == ErrorReason_DB_UPDATEERR.String() && e.Code == 300
}

func ErrorDbUpdateerr(format string, args ...interface{}) *errors.Error {
	return errors.New(300, ErrorReason_DB_UPDATEERR.String(), fmt.Sprintf(format, args...))
}

func IsParamErr(err error) bool {
	if err == nil {
		return false
	}
	e := errors.FromError(err)
	return e.Reason == ErrorReason_Param_Err.String() && e.Code == 301
}

func ErrorParamErr(format string, args ...interface{}) *errors.Error {
	return errors.New(301, ErrorReason_Param_Err.String(), fmt.Sprintf(format, args...))
}

func IsDbSaveerror(err error) bool {
	if err == nil {
		return false
	}
	e := errors.FromError(err)
	return e.Reason == ErrorReason_DB_SAVEERROR.String() && e.Code == 302
}

func ErrorDbSaveerror(format string, args ...interface{}) *errors.Error {
	return errors.New(302, ErrorReason_DB_SAVEERROR.String(), fmt.Sprintf(format, args...))
}

func IsDbDeleteerror(err error) bool {
	if err == nil {
		return false
	}
	e := errors.FromError(err)
	return e.Reason == ErrorReason_DB_DELETEERROR.String() && e.Code == 303
}

func ErrorDbDeleteerror(format string, args ...interface{}) *errors.Error {
	return errors.New(303, ErrorReason_DB_DELETEERROR.String(), fmt.Sprintf(format, args...))
}

func IsCrawlerError(err error) bool {
	if err == nil {
		return false
	}
	e := errors.FromError(err)
	return e.Reason == ErrorReason_Crawler_Error.String() && e.Code == 304
}

func ErrorCrawlerError(format string, args ...interface{}) *errors.Error {
	return errors.New(304, ErrorReason_Crawler_Error.String(), fmt.Sprintf(format, args...))
}

func IsCCNULoginError(err error) bool {
	if err == nil {
		return false
	}
	e := errors.FromError(err)
	return e.Reason == ErrorReason_CCNULogin_Error.String() && e.Code == 305
}

func ErrorCCNULoginError(format string, args ...interface{}) *errors.Error {
	return errors.New(305, ErrorReason_CCNULogin_Error.String(), fmt.Sprintf(format, args...))
}

func IsScidnotexistErroe(err error) bool {
	if err == nil {
		return false
	}
	e := errors.FromError(err)
	return e.Reason == ErrorReason_SCIDNOTEXIST_Erroe.String() && e.Code == 306
}

func ErrorScidnotexistErroe(format string, args ...interface{}) *errors.Error {
	return errors.New(306, ErrorReason_SCIDNOTEXIST_Erroe.String(), fmt.Sprintf(format, args...))
}

func IsRecyclebindonothavetheclass(err error) bool {
	if err == nil {
		return false
	}
	e := errors.FromError(err)
	return e.Reason == ErrorReason_RECYCLEBINDONOTHAVETHECLASS.String() && e.Code == 307
}

func ErrorRecyclebindonothavetheclass(format string, args ...interface{}) *errors.Error {
	return errors.New(307, ErrorReason_RECYCLEBINDONOTHAVETHECLASS.String(), fmt.Sprintf(format, args...))
}

func IsRecoverfailed(err error) bool {
	if err == nil {
		return false
	}
	e := errors.FromError(err)
	return e.Reason == ErrorReason_RECOVERFAILED.String() && e.Code == 308
}

func ErrorRecoverfailed(format string, args ...interface{}) *errors.Error {
	return errors.New(308, ErrorReason_RECOVERFAILED.String(), fmt.Sprintf(format, args...))
}
