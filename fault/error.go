// Package makross is a high productive and modular web framework in Golang.

// Package fault provides a panic and error handler for the makross.
package fault

import (
	"github.com/insionng/makross"
)

// ErrorHandler returns a handler that handles errors returned by the handlers following this one.
// If the error implements makross.HTTPError, the handler will set the HTTP status code accordingly.
// Otherwise the HTTP status is set as http.StatusInternalServerError. The handler will also write the error
// as the response body.
//
// A log function can be provided to log a message whenever an error is handled. If nil, no message will be logged.
//
// An optional error conversion function can also be provided to convert an error into a normalized one
// before sending it to the response.
//
//     import (
//         "log"
//         "github.com/insionng/makross"
//         "github.com/insionng/makross/fault"
//     )
//
//     r := makross.New()
//     r.Use(fault.ErrorHandler(log.Printf))
//     r.Use(fault.PanicHandler(log.Printf))
func ErrorHandler(logf LogFunc, errorf ...ConvertErrorFunc) makross.Handler {
	return func(c *makross.Context) error {
		err := c.Next()
		if err == nil {
			return nil
		}

		if logf != nil {
			logf("%v", err)
		}

		if len(errorf) > 0 {
			err = errorf[0](c, err)
		}

		c.HandleError(err)
		c.Abort()

		return nil
	}
}
