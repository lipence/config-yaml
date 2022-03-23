package yaml

import (
	"context"
	"fmt"
	"reflect"
	"strconv"

	"github.com/lipence/config"
	"github.com/lipence/config/utils"
	"github.com/lipence/gabs-yaml/v2"
	"gopkg.in/yaml.v3"
)

type listIterator struct {
	offset int
	data   []*gabs.Container
}

func (l *listIterator) Next() bool {
	if l.offset >= len(l.data)-1 {
		return false
	}
	l.offset++
	return true
}

func (l *listIterator) Value() config.Value {
	return &value{result: l.data[l.offset]}
}

func (l *listIterator) Label() string {
	return strconv.FormatInt(int64(l.offset), 10)
}

type structIterator struct {
	offset int
	keys   []string
	data   map[string]*gabs.Container
}

func (s *structIterator) Next() bool {
	if s.offset >= len(s.keys)-1 {
		return false
	}
	s.offset++
	return true
}

func (s *structIterator) Value() config.Value {
	return &value{result: s.data[s.keys[s.offset]]}
}

func (s *structIterator) Label() string {
	return s.keys[s.offset]
}

func NewJsonValue(c *gabs.Container) *value {
	return &value{result: c}
}

func Parse(content []byte) (*value, error) {
	if c, err := gabs.ParseYAML(content); err != nil {
		return nil, err
	} else {
		return NewJsonValue(c), nil
	}
}

type value struct {
	result *gabs.Container
}

type ctxBypass struct {
	context.Context
}

func (*ctxBypass) __bypass() {}

func (v *value) Decode(target interface{}) error {
	return v.DecodeWithCtx(&ctxBypass{context.Background()}, target)
}

func (v *value) DecodeWithCtx(ctx context.Context, target interface{}) (err error) {
	if decoder, ok := target.(config.Decoder); ok {
		if err = decoder.Decode(v); err != nil {
			return fmt.Errorf("%w: (position: %s)", err, v.Ref())
		}
		return nil
	}
	if decoder, ok := target.(config.CtxDecoder); ok {
		var _bypass *ctxBypass
		if _bypass, ok = ctx.(*ctxBypass); ok {
			ctx = _bypass.Context
		}
		if err = decoder.Decode(ctx, v); err != nil {
			return fmt.Errorf("%w: (position: %s)", err, v.Ref())
		}
		return nil
	}
	if decoder, ok := target.(config.CtxConfigDecoder); ok {
		var _bypass *ctxBypass
		if _bypass, ok = ctx.(*ctxBypass); ok {
			ctx = _bypass.Context
		}
		if err = decoder.DecodeConfig(ctx, v); err != nil {
			return fmt.Errorf("%w: (position: %s)", err, v.Ref())
		}
		return nil
	}
	var jBytes []byte
	if jBytes, err = v.result.MarshalYAML(); err != nil {
		return fmt.Errorf("%w: (position: %s)", err, v.Ref())
	}
	return yaml.Unmarshal(jBytes, target)
}

func (v *value) String() (string, error) {
	return utils.ItfToString(v.result.Data())
}

func (v *value) StringList() ([]string, error) {
	data := v.result.Data()
	return utils.ItfToStringSlice(data)
}

func (v *value) Bytes() ([]byte, error) {
	return utils.ItfToBytes(v.result.Data())
}

func (v *value) Bool() (bool, error) {
	return utils.ItfToBoolean(v.result.Data())
}

func (v *value) Float64() (float64, error) {
	return utils.ItfToFloat64(v.result.Data())
}

func (v *value) Int64() (int64, error) {
	return utils.ItfToInt64(v.result.Data())
}

func (v *value) Uint64() (uint64, error) {
	return utils.ItfToUInt64(v.result.Data())
}

func (v *value) Interface() (interface{}, error) {
	return v.result.Data(), nil
}

func (v *value) Ref() string {
	return "unsupported operation"
}

func (v *value) File() string {
	return "/tmp"
}

func (v *value) Lookup(path ...string) (config.Value, bool) {
	newResult := v.result.Search(path...)
	if newResult == nil {
		return nil, false
	}
	return &value{result: newResult}, true
}

func (v *value) List() (config.Iterator, error) {
	var data = v.result.Data()
	if _, ok := data.([]interface{}); !ok {
		return nil, fmt.Errorf("unsupported iterator type `%s`", reflect.TypeOf(data))
	} else {
		return &listIterator{data: v.result.Children(), offset: -1}, nil
	}
}

func (v *value) Struct() (config.Iterator, error) {
	var data = v.result.Data()
	if _, ok := data.(map[string]interface{}); !ok {
		return nil, fmt.Errorf("unsupported iterator type `%s`", reflect.TypeOf(data))
	} else {
		dataMap := v.result.ChildrenMap()
		dataKeys := make([]string, 0, len(dataMap))
		for k := range dataMap {
			dataKeys = append(dataKeys, k)
		}
		return &structIterator{data: dataMap, offset: -1, keys: dataKeys}, nil
	}
}

func (v *value) Kind() config.Kind {
	data, err := v.Interface()
	if err != nil {
		return config.UndefinedKind
	}
	if data == nil {
		return config.NullKind
	}
	dataType := reflect.TypeOf(data)
	switch dataType.Kind() {
	default:
		fallthrough
	case reflect.Invalid, reflect.UnsafePointer, reflect.Interface, reflect.Chan, reflect.Func:
		return config.UndefinedKind
	case reflect.String:
		return config.StringKind
	case reflect.Bool:
		return config.BoolKind
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return config.NumberKind
	case reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128:
		return config.DecimalKind
	case reflect.Struct, reflect.Map:
		return config.StructKind
	case reflect.Slice, reflect.Array:
		if dataType.Elem() == reflect.TypeOf(byte(0)) {
			return config.BytesKind
		}
		return config.ListKind
	}
}

func (v *value) Marshal() ([]byte, error) {
	return v.result.MarshalYAML()
}
