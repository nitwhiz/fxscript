package vm

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/nitwhiz/fxscript/fx"
)

var ErrInvalidOptional = fmt.Errorf("argument type cannot be optional")

type MissingArgumentError struct {
	Index    int
	Name     string
	TypeName string
}

func (e *MissingArgumentError) Error() string {
	return fmt.Sprintf("missing argument value at index %d: '%s' (%s)", e.Index, e.Name, e.TypeName)
}

type ArgumentTypeError struct {
	Index    int
	Name     string
	TypeName string
	Err      error
}

func (e *ArgumentTypeError) Error() string {
	return fmt.Sprintf("invalid argument at index %d: '%s' (%s): %v", e.Index, e.Name, e.TypeName, e.Err)
}

func (e *ArgumentTypeError) Unwrap() error {
	return e.Err
}

func (f *RuntimeFrame) unmarshalArgs(argv []fx.ExpressionNode, v any) error {
	typ := reflect.TypeOf(v).Elem()
	val := reflect.ValueOf(v).Elem()

	for i := 0; i < typ.NumField(); i++ {
		typField := typ.Field(i)
		argTag := typField.Tag.Get("arg")

		if argTag != "-" {
			segments := strings.Split(argTag, ",")

			var argIdx int

			if segments[0] == "" {
				argIdx = i
			} else {
				var err error

				argIdx, err = strconv.Atoi(segments[0])

				if err != nil {
					return err
				}
			}

			useDefaultValue := false

			if len(argv) <= argIdx {
				if len(segments) == 2 && segments[1] == "optional" {
					useDefaultValue = true
				} else {
					return &MissingArgumentError{argIdx, typField.Name, typField.Type.Name()}
				}
			}

			valField := val.Field(i)

			if useDefaultValue {
				switch valField.Interface().(type) {
				case fx.Identifier:
					return &ArgumentTypeError{argIdx, typField.Name, typField.Type.Name(), ErrInvalidOptional}
				}
			} else {
				node := argv[argIdx]
				rawValue, err := f.Eval(node)

				if err != nil {
					return err
				}

				switch valField.Interface().(type) {
				case fx.Identifier:
					if identNode, ok := node.(*fx.IdentifierNode); ok {
						valField.Set(reflect.ValueOf(identNode.Identifier))
					} else {
						switch v := rawValue.(type) {
						case int:
							valField.Set(reflect.ValueOf(fx.Identifier(v)))
							break
						case float64:
							valField.Set(reflect.ValueOf(fx.Identifier(int(v))))
							break
						default:
							return &ArgumentTypeError{argIdx, typField.Name, typField.Type.Name(), fmt.Errorf("unsupported type: %T", rawValue)}
						}
					}
					break
				case int:
					switch numericValue := rawValue.(type) {
					case int:
						valField.Set(reflect.ValueOf(numericValue))
						break
					case float64:
						valField.Set(reflect.ValueOf(int(numericValue)))
						break
					}

					break
				case float64:
					switch numericValue := rawValue.(type) {
					case float64:
						valField.Set(reflect.ValueOf(numericValue))
						break
					case int:
						valField.Set(reflect.ValueOf(float64(numericValue)))
						break
					}

					break
				case string:
					valField.Set(reflect.ValueOf(rawValue))
					break
				default:
					return &ArgumentTypeError{argIdx, typField.Name, typField.Type.Name(), fmt.Errorf("unsupported type: %T", rawValue)}
				}
			}
		}
	}

	return nil
}

func WithArgs[ArgsType any](f *RuntimeFrame, cmdArgs []fx.ExpressionNode, h func(f *RuntimeFrame, args *ArgsType) (jumpTarget int, jump bool)) (jumpTarget int, jump bool) {
	args := new(ArgsType)

	if err := f.unmarshalArgs(cmdArgs, args); err != nil {
		f.HandleError(err)
	}

	return h(f, args)
}
