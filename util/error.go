package util

import (
	. "fmt"
)

//Errors

type HrootError struct {
	cause error
	message string
}

//Returns nested error
func (err HrootError) Cause() error {
	return err.cause
}

//Golang stdlib func
func (err HrootError) Error() string {
	return err.message
}

//Sugar
func ExitGently(a ...interface{}) {
	panic(HrootError{message: Sprintln(a...)})
}
