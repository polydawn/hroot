package util

import (
	. "fmt"
)

//Errors

type DocketError struct {
	cause error
	message string
}

//Returns nested error
func (err DocketError) Cause() error {
	return err.cause
}

//Golang stdlib func
func (err DocketError) Error() string {
	return err.message
}

//Sugar
func ExitGently(a ...interface{}) {
	panic(DocketError{message: Sprintln(a...)})
}
