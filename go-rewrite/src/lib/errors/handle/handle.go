package handle

import "github.com/cockroachdb/errors"

func Wrap(handledErr error, newMarkingError error, msg string) error {
	newErr := errors.Mark(handledErr, newMarkingError)
	return errors.Wrap(newErr, msg)
}

func Message(newMarkingError error, msg string) error {
	err := errors.New(msg)
	return errors.Mark(err, newMarkingError)
}
