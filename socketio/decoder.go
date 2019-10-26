package socketio

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"

	"github.com/golang/glog"
)

// SocketIO packet types.
type PacketType int

// Defined in: https://preview.tinyurl.com/yxcgen7t
const (
	Open PacketType = iota
	Close
	Ping
	Pong
	Message
	Upgrade
	Noop
)

// SocketIO event types.
type MessageType int

// Most are unused. Defined in: https://preview.tinyurl.com/y3s4eh2y
const (
	Connect MessageType = iota
	Disconnect
	Event
	Ack
	Error
	BinaryEvent
	BinaryAck
)

// The general SocketIO packet metadata.
type PacketData struct {
	PacketType  PacketType
	MessageType MessageType
	Namespace   string
	// The JSON string data remaining after the metadata has been parsed
	// out.
	Data string
}

// SocketIO data uses a format <length>:<data>. This function splits on the
// first occurrence of ":", attempts to parse <length> as an int, and returns
// <data>. If there is a problem, the original string is returned. The method
// returns a second string parameter containing the remainder of the string if
// any.
func splitOnLength(input string) (string, string) {
	parts := strings.SplitN(input, ":", 2)
	if len(parts) != 2 {
		return input, ""
	}
	length, err := strconv.Atoi(parts[0])
	if err != nil {
		if glog.V(5) {
			glog.Warningf("%s is not a length", parts[0])
		}
		return input, ""
	}
	if glog.V(5) {
		glog.Infof("Found response of length %d", length)
		glog.Infof("Length actual data is %d", len(parts[1]))
	}
	return parts[1][:length], parts[1][length:]
}

// Returns true if the first character is a number and sets the <name> field of
// the passed in interface to the retrieved value if it exists. Also, the first
// char is removed from the decoder. Returns false if the first char is not a
// number.
func maybeProcessFirstChar(
	name string, data string, v interface{}) bool {
	firstChar := data[0]
	number, err := strconv.Atoi(string(firstChar))
	if err != nil {
		if glog.V(3) {
			glog.Warningf("No %s found", name)
		}
		return false
	}
	instance := reflect.ValueOf(v).Elem()
	typeOfV := instance.Type()
	for i := 0; i < instance.NumField(); i++ {
		f := instance.Field(i)
		if typeOfV.Field(i).Name == name && f.Kind() == reflect.Int {
			if glog.V(3) {
				glog.Infof(
					"Setting %s to %d",
					name, number)
			}
			f.SetInt(int64(number))
		}
	}
	return true
}

// Given a string of data, this method will attempt to parse out a namespace
// prefix. If it finds one and the passed in interface has a Namespace field,
// this method will set the field to the parsed value. Returns the original
// string if no namespace was found. Otherwise, the remaining string data is
// returned.
func maybeProcessNamespace(data string, v interface{}) string {
	firstComma := strings.Index(data, ",")
	firstOpenBracket := strings.Index(data, "[")
	if data[0] == '/' && firstComma > -1 && firstComma < firstOpenBracket {
		parts := strings.SplitN(data, ",", 2)
		if glog.V(3) {
			glog.Infof("Found namespace: %s", parts[0])
		}
		instance := reflect.ValueOf(v).Elem()
		typeOfV := instance.Type()
		for i := 0; i < instance.NumField(); i++ {
			f := instance.Field(i)
			if typeOfV.Field(i).Name == "Namespace" &&
				f.Kind() == reflect.String {
				if glog.V(3) {
					glog.Infof(
						"Setting Namespace to %s",
						parts[0])
				}
				f.SetString(parts[0])
				return parts[1]
			}
		}
		return parts[1]
	}
	return data
}

// An error type used when a potential JSON string is invalid.
type NotJsonError struct {
	data string
}

func (n *NotJsonError) Error() string {
	return n.data
}

// Parses the PacketType, MessageType and Namespace out of the passed in data
// string and into the fields of the same name on the passed in type v. The
// remaining string data is returned. If the metadata cannot be found or the
// passed in type v does not have PacketType, MessageType or Namespace fields,
// then no changes are made and the original data is returned.
func ParseMetadata(data string, v interface{}) string {
	minusTypes := data
	if maybeProcessFirstChar("PacketType", minusTypes, v) {
		minusTypes = minusTypes[1:]
		if maybeProcessFirstChar("MessageType", minusTypes, v) {
			minusTypes = minusTypes[1:]
		}
	}
	if len(minusTypes) == 0 {
		return ""
	}
	return maybeProcessNamespace(minusTypes, v)
}

// Parses the actual JSON message into the passed in message type. The SocketIO
// response seems to alternate between a JSON array and a JSON object. In the
// case of the former, this method attempts to parse the second element of the
// array into v. If an error occurs, it is returned and v may not contain all
// parsed data.
func ParseToJSON(data string, v interface{}) error {
	// The resulting is either a JSON object or a JSON array starting with
	// the SocketIO event type string and ending with the JSON object. In
	// order to handle both of these scenarios, Unmarshal first tries to
	// parse a JSON array with the first element being an instance of v. If
	// that fails, it tries to parse all the data into v.
	if glog.V(5) {
		glog.Infof("Checking JSON validity of %s", string(data))
	}
	if !json.Valid([]byte(data)) {
		return &NotJsonError{"invalid JSON"}
	}
	// Sometimes, the JSON is an array containing a string event type
	// followed by the JSON object. Othertimes, it is just the object. Use
	// jsonArray to test for the first case.
	var jsonArray []json.RawMessage
	err := json.Unmarshal([]byte(data), &jsonArray)
	if err != nil {
		if glog.V(3) {
			glog.Warningf(
				"Could not parse response as JSON array: %s",
				err)
		}
		return json.Unmarshal([]byte(data), v)
	}
	if glog.V(3) {
		glog.Infof("Parsed as JSON array: %s", string(jsonArray[1]))
	}
	jsonPart := jsonArray[1]
	err = json.Unmarshal(jsonPart, v)
	if err != nil {
		if glog.V(3) {
			glog.Errorf("Could not unmarshal data: %s", err)
		}
		jsonPart, err := strconv.Unquote(string(jsonPart))
		if err != nil {
			glog.Error("Could not unescape JSON string")
		}
		return json.Unmarshal([]byte(jsonPart), v)
	}
	return nil
}

// Parses the JSON HTTP SocketIO response from the given Reader into the passed
// in structs. For each of the passed in structs, if they contain MessageType
// or PacketType fields of type int, those fields will be populated with the
// corresponding response values.
func HTTPToJSON(data io.Reader, v []interface{}) error {
	bytes, err := ioutil.ReadAll(data)
	if err != nil {
		glog.Errorf("Could not read input data: %s", err)
	}
	response := string(bytes)
	glog.Infof("Parsing HTTP Response: %s", response)

	fillingIn := 0
	for true {
		data, leftover := splitOnLength(response)
		if glog.V(3) {
			glog.Infof("Subresponse: %s", data)
			glog.Infof("Leftover: %s", leftover)
		}
		remaining := ParseMetadata(data, v[fillingIn])
		if len(remaining) > 0 {
			err := ParseToJSON(remaining, v[fillingIn])
			if err != nil {
				glog.Warningf(
					"Unable to parse message: %s; %s",
					data, err)
				return err
			}
		}
		if len(leftover) == 0 {
			break
		}
		response = leftover
		fillingIn++
	}
	return nil
}
