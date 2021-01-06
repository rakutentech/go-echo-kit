package errors

import (
	"errors"
	"reflect"
	"testing"
)

func TestNewErrorWithMsg(t *testing.T) {
	var testErrCode ErrorCode = "hoge"
	var testMsg = "fuga fuga"

	var want = Error{errors.New("fuga fuga"), "hoge"}
	var got = NewErrorWithMsg(testErrCode, testMsg)

	if !reflect.DeepEqual(want, got) {
		t.Errorf("want: %v but got: %v", want, got)
	}
}

// NewErrorWithMsgf - creates an Error instance with formatted message
func TestNewErrorWithMsgf(t *testing.T) {
	var testErrCode ErrorCode = "hoge"
	var testFormat = "%s, %s, %s"
	var testValue1, testValue2, testValue3 = "fuga1", "fuga2", "fuga3"

	var want = Error{errors.New("fuga1, fuga2, fuga3"), "hoge"}
	var got = NewErrorWithMsgf(testErrCode, testFormat, testValue1, testValue2, testValue3)

	if !reflect.DeepEqual(want, got) {
		t.Errorf("want: %v but got: %v", want, got)
	}
}
