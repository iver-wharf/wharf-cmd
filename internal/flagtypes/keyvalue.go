package flagtypes

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/pflag"
)

// ensure they conform to the interfaces
var _ pflag.Value = &KeyValue{}
var _ pflag.Value = &KeyValueArray{}
var _ pflag.SliceValue = &KeyValueArray{}

// KeyValue is a flag that takes in a key=value format.
type KeyValue struct {
	Key   string
	Value string
}

// String returns a "key=value" string represetntation for this flag.
func (s *KeyValue) String() string {
	return fmt.Sprintf("%s=%s", s.Key, s.Value)
}

// Set parses a "key=value" string and updates this flag.
func (s *KeyValue) Set(val string) error {
	key, value, ok := strings.Cut(val, "=")
	if !ok {
		return errors.New("missing delimiter \"=\"")
	}
	if key == "" {
		return errors.New("empty key")
	}
	s.Key = key
	s.Value = value
	return nil
}

// Type returns the name of this type.
func (s *KeyValue) Type() string {
	return "keyvalue"
}

// KeyValueArray is a flag type that takes in key=value on each flag value and
// can be specified multiple times.
type KeyValueArray struct {
	Pairs   []KeyValue
	changed bool
}

// String returns a "[key1=value,key2=value]" string represetntation for this flag.
func (s *KeyValueArray) String() string {
	str, _ := writeAsCSV(s.GetSlice())
	return fmt.Sprintf("[%s]", str)
}

// Set parses a "key=value" string and updates this flag with that value,
// overriding the array if it's the first time its set, or appends the value
// if it's a consecutive call.
func (s *KeyValueArray) Set(val string) error {
	var kv KeyValue
	if err := kv.Set(val); err != nil {
		return err
	}
	if !s.changed {
		s.Pairs = []KeyValue{kv}
		s.changed = true
	} else {
		s.Pairs = append(s.Pairs, kv)
	}
	return nil
}

// Type returns the name of this type.
func (s *KeyValueArray) Type() string {
	return "keyvalues"
}

// Append parses a "key=value" string and adds that to this array.
func (s *KeyValueArray) Append(val string) error {
	var kv KeyValue
	if err := kv.Set(val); err != nil {
		return err
	}
	s.Pairs = append(s.Pairs, kv)
	return nil
}

// Replace parses a slice of "key=value" strings and sets those as this flag's
// new values.
func (s *KeyValueArray) Replace(val []string) error {
	out := make([]KeyValue, len(val))
	for i, d := range val {
		var kv KeyValue
		if err := kv.Set(d); err != nil {
			return err
		}
		out[i] = kv
	}
	s.Pairs = out
	return nil
}

// GetSlice returns a slice of "key=value" string representations for all the
// values.
func (s *KeyValueArray) GetSlice() []string {
	out := make([]string, len(s.Pairs))
	for i, d := range s.Pairs {
		out[i] = d.String()
	}
	return out
}

func writeAsCSV(vals []string) (string, error) {
	b := &bytes.Buffer{}
	w := csv.NewWriter(b)
	err := w.Write(vals)
	if err != nil {
		return "", err
	}
	w.Flush()
	return strings.TrimSuffix(b.String(), "\n"), nil
}
