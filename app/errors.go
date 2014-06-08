package app

import (
	"github.com/ngerakines/codederror"
	"log"
)

var (
	ErrorNotImplemented = codederror.NewCodedError([]string{"TRM", "APP"}, 1, "Something wasn't implemented")

	AllErrors = []codederror.CodedError{
		ErrorNotImplemented,
	}
)

// DumpErrors prints out all of the errors contained in AllErrors.
func DumpErrors() {
	for _, err := range AllErrors {
		log.Println(err.Error(), err.Namespaces(), err.Code(), err.Description())
	}
}
