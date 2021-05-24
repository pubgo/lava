package errors

import "errors"

// IsUnavailable determines if err is an error which indicates a unavailable error.
// It supports wrapped errors.
func IsUnavailable(err error) bool {
	if se := new(Error); errors.As(err, &se) {
		return se.Code == 14
	}
	return false
}

// IsDataLoss determines if err is an error which indicates a data loss error.
// It supports wrapped errors.
func IsDataLoss(err error) bool {
	if se := new(Error); errors.As(err, &se) {
		return se.Code == 15
	}
	return false
}

// IsUnauthorized determines if err is an error which indicates a unauthorized error.
// It supports wrapped errors.
func IsUnauthorized(err error) bool {
	if se := new(Error); errors.As(err, &se) {
		return se.Code == 16
	}
	return false
}

// IsCancelled determines if err is an error which indicates a cancelled error.
// It supports wrapped errors.
func IsCancelled(err error) bool {
	if se := new(Error); errors.As(err, &se) {
		return se.Code == 1
	}
	return false
}

// IsUnknown determines if err is an error which indicates a unknown error.
// It supports wrapped errors.
func IsUnknown(err error) bool {
	if se := new(Error); errors.As(err, &se) {
		return se.Code == 2
	}
	return false
}

// IsInvalidArgument determines if err is an error which indicates an invalid argument error.
// It supports wrapped errors.
func IsInvalidArgument(err error) bool {
	if se := new(Error); errors.As(err, &se) {
		return se.Code == 3
	}
	return false
}

// IsDeadlineExceeded determines if err is an error which indicates a deadline exceeded error.
// It supports wrapped errors.
func IsDeadlineExceeded(err error) bool {
	if se := new(Error); errors.As(err, &se) {
		return se.Code == 4
	}
	return false
}

// IsNotFound determines if err is an error which indicates a not found error.
// It supports wrapped errors.
func IsNotFound(err error) bool {
	if se := new(Error); errors.As(err, &se) {
		return se.Code == 5
	}
	return false
}

// IsAlreadyExists determines if err is an error which indicates a already exsits error.
// It supports wrapped errors.
func IsAlreadyExists(err error) bool {
	if se := new(Error); errors.As(err, &se) {
		return se.Code == 6
	}
	return false
}

// IsPermissionDenied determines if err is an error which indicates a permission denied error.
// It supports wrapped errors.
func IsPermissionDenied(err error) bool {
	if se := new(Error); errors.As(err, &se) {
		return se.Code == 7
	}
	return false
}

// IsResourceExhausted determines if err is an error which indicates a resource exhausted error.
// It supports wrapped errors.
func IsResourceExhausted(err error) bool {
	if se := new(Error); errors.As(err, &se) {
		return se.Code == 8
	}
	return false
}

// IsFailedPrecondition determines if err is an error which indicates a failed precondition error.
// It supports wrapped errors.
func IsFailedPrecondition(err error) bool {
	if se := new(Error); errors.As(err, &se) {
		return se.Code == 9
	}
	return false
}

// IsAborted determines if err is an error which indicates an aborted error.
// It supports wrapped errors.
func IsAborted(err error) bool {
	if se := new(Error); errors.As(err, &se) {
		return se.Code == 10
	}
	return false
}

// IsOutOfRange determines if err is an error which indicates a out of range error.
// It supports wrapped errors.
func IsOutOfRange(err error) bool {
	if se := new(Error); errors.As(err, &se) {
		return se.Code == 11
	}
	return false
}

// IsUnimplemented determines if err is an error which indicates a unimplemented error.
// It supports wrapped errors.
func IsUnimplemented(err error) bool {
	if se := new(Error); errors.As(err, &se) {
		return se.Code == 12
	}
	return false
}

// IsInternal determines if err is an error which indicates an internal server error.
// It supports wrapped errors.
func IsInternal(err error) bool {
	if se := new(Error); errors.As(err, &se) {
		return se.Code == 13
	}
	return false
}
