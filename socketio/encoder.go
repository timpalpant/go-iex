package socketio

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"reflect"
	"strings"

	"github.com/golang/glog"
)

var disallowedTypes = map[reflect.Kind]struct{}{
	reflect.Chan:          {},
	reflect.Func:          {},
	reflect.Interface:     {},
	reflect.Map:           {},
	reflect.Ptr:           {},
	reflect.Struct:        {},
	reflect.UnsafePointer: {},
}

// Encodes messages for use with IEX SocketIO. MessageType and PacketType are
// defined in decoder.go. The values of the fields of the passed in interface
// are converted into a JSON array of strings. If the field value is an array,
// the elements are converted into a single, comma-joined string before being
// added to the resulting JSON string array. If the MessageType or PacketType
// are less than 0, they are not set on the output.
type Encoder interface {
	Encode(m MessageType, p PacketType, v interface{}) (io.Reader, error)
}

// Wraps a strArrayEncoder and returns its contents prepended by <length>:.
type httpEncoder struct {
	content *strArrayEncoder
}

func (enc *httpEncoder) Encode(
	m MessageType, p PacketType, v interface{}) (io.Reader, error) {
	inner, err := enc.content.Encode(m, p, v)
	if err != nil {
		return nil, err
	}
	val, err := ioutil.ReadAll(inner)
	if err != nil {
		if glog.V(3) {
			glog.Warningf("Failed to read inner encoding: %q", err)
		}
		return nil, err
	}
	if glog.V(3) {
		glog.Infof("Inner encoding: %s", val)
	}
	parts := []string{fmt.Sprintf("%d", len(val)), string(val)}
	return strings.NewReader(strings.Join(parts, ":")), nil
}

// The base encoder implementation that performs as described by the interface.
type strArrayEncoder struct {
	namespace string
}

// Used to indicate an encoding error.
type encodeError struct {
	message string
}

func (e *encodeError) Error() string {
	return e.message
}

func (enc *strArrayEncoder) Encode(
	m MessageType, p PacketType, v interface{}) (io.Reader, error) {
	readers := make([]io.Reader, 0)
	if m >= 0 {
		readers = append(readers,
			strings.NewReader(fmt.Sprintf("%d", m)))
	}
	if p >= 0 {
		readers = append(readers,
			strings.NewReader(fmt.Sprintf("%d", p)))
	}
	if len(enc.namespace) > 0 {
		readers = append(readers,
			strings.NewReader(enc.namespace+","))
	}
	if v == nil {
		return io.MultiReader(readers...), nil
	}
	parts := make([]string, 0)
	instance := reflect.ValueOf(v).Elem()
	// For each of the fields in v, turn the value into a string and append
	// it to parts. If the field is of type Array, join the elements using
	// commas and then add the resulting string to parts. Complex types
	// other than Array or Slice cannot be converted and will result in an
	// error.
	for i := 0; i < instance.NumField(); i++ {
		field := instance.Field(i)
		// Skip unset fields.
		if field.IsZero() {
			if glog.V(3) {
				glog.Infof("Skipping unset field %s",
					instance.Type().Field(i).Name)
			}
			continue
		}
		kind := field.Kind()
		_, disallowed := disallowedTypes[kind]
		if disallowed {
			return nil, &encodeError{fmt.Sprintf(
				"Cannot encode type: %s", field.Type())}
		}
		if glog.V(3) {
			glog.Infof("Encoding %s", field.String())
		}
		if kind != reflect.Array && kind != reflect.Slice {
			strEncoding := fmt.Sprintf("%v", field.Interface())
			if len(strEncoding) > 0 {
				parts = append(parts, strEncoding)
			}
		}
		if kind == reflect.Array || kind == reflect.Slice {
			elemType := field.Type().Elem().Kind()
			_, disallowed = disallowedTypes[elemType]
			if disallowed {
				return nil, &encodeError{fmt.Sprintf(
					"Cannot encode Array type: %s",
					field.Type())}
			}
			subParts := make([]string, 0)
			for j := 0; j < field.Len(); j++ {
				strEncoding := fmt.Sprintf(
					"%v", field.Index(j).Interface())
				if len(strEncoding) > 0 {
					subParts = append(subParts, strEncoding)
				}
			}
			parts = append(parts, strings.Join(subParts, ","))
		}
	}
	if glog.V(3) {
		glog.Infof("Encoding parts: %v", parts)
	}
	encoding, err := json.Marshal(parts)
	if err != nil {
		glog.Errorf("Failed to encode data as JSON: %s", err)
		return nil, err
	}
	if len(parts) > 0 {
		readers = append(readers, bytes.NewBuffer(encoding))
	}
	return io.MultiReader(readers...), nil

}

// Returns an encoder for use with HTTP Post.
func NewHTTPEncoder(namespace string) Encoder {
	return &httpEncoder{&strArrayEncoder{namespace}}
}

// Returns an encoder for use with SocketIO.
func NewWSEncoder(namespace string) Encoder {
	return &strArrayEncoder{namespace}
}
