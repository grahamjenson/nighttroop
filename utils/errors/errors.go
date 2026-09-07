package errors

import (
	"fmt"
	"log/slog"
	"net/http"
)

func HandleInternalError(w http.ResponseWriter, r *http.Request, statusCode int, message string, err error, args ...any) {
	allArgs := append([]any{
		"error", err,
		"stack", localStack(),
	}, args...)
	slog.Info(message, allArgs...)
	http.Error(w, "Internal Error", statusCode)
}

func DumpErrorChain(err error) {
	if err == nil {
		fmt.Println("error is nil")
		return
	}

	fmt.Println("ERROR CHAIN:")

	for i := 0; err != nil; i++ {
		fmt.Printf("[%d] type=%T\n", i, err)
		fmt.Printf("    value=%v\n", err)

		// Check whether this error implements Unwrap() error.
		unwrapper, ok := err.(interface {
			Unwrap() error
		})

		if !ok {
			fmt.Println("    no Unwrap()")
			break
		}

		err = unwrapper.Unwrap()
	}
}
