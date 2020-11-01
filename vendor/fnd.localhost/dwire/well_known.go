package dwire

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

// Encodes a time.Time value into the Writer. Since time.Time is a
// well-known type, you likely do not need to call this method directly.
// Instead, provide the time value to EncodeField or EncodeFields.
func EncodeTime(w io.Writer, val interface{}) error {
	cast, ok := val.(time.Time)
	if !ok {
		return errors.New("value is not a time.Time")
	}

	unix := cast.Unix()
	if unix < 0 {
		return errors.New("negative UNIX time")
	}
	return EncodeField(w, uint32(unix))
}

// Decodes a time.Time value from the Reader. Since time.Time is a
// well-known type, you likely do not need to call this method directly.
// Instead, call EncodeField or EncodeFields with a &time.Time item.
func DecodeTime(r io.Reader, val interface{}) error {
	cast, ok := val.(*time.Time)
	if !ok {
		return errors.New("value is not a *time.Time")
	}

	var unixTs uint32
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
