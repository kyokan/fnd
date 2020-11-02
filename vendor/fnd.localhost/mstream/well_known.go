package mstream

import (
	"fmt"
	"io"
	"reflect"
	"time"

	"github.com/pkg/errors"
)

type encoderFunc func(w io.Writer, val interface{}) error
type decoderFunc func(r io.Reader, val interface{}) error

var (
	wellKnownEncoders = make(map[string]encoderFunc)
	wellKnownDecoders = make(map[string]decoderFunc)
)

func EncodeTime(w io.Writer, val interface{}) error {
	cast, ok := val.(time.Time)
	if !ok {
		return errors.New("value is not a time.Time")
	}

	return EncodeField(w, uint64(cast.Unix()))
}

func DecodeTime(r io.Reader, val interface{}) error {
	cast, ok := val.(*time.Time)
	if !ok {
		return errors.New("value is not a *time.Time")
	}

	var unixTs uint64
	if err := DecodeField(r, &unixTs); err != nil {
		return errors.Wrap(err, "failed to decode timestamp")
	}

	*cast = time.Unix(int64(unixTs), 0)
	return nil
}

func canonicalizeWellKnown(t reflect.Type) string {
	return fmt.Sprintf("%s/%s", t.PkgPath(), t.Name())
}

func init() {
	timeTypeKey := canonicalizeWellKnown(reflect.TypeOf(time.Time{}))
	wellKnownEncoders[timeTypeKey] = EncodeTime
	wellKnownDecoders[timeTypeKey] = DecodeTime
}
