package faunadb

import (
	"reflect"
	"strings"
)

func printFn(fn interface{}) string {

	t := reflect.TypeOf(fn)
	if t.Kind() != reflect.Struct {
		return "fnApply should only be of type struct"
	}

	var name string
	var sbOpt strings.Builder
	var sbArgs strings.Builder
	val := reflect.ValueOf(fn)

	for idx := 0; idx < t.NumField(); idx++ {
		field := t.Field(idx)
		if field.Name == "fnApply" {
			continue
		}
		if idx == 1 {
			name = field.Name
		}

		fv := val.Field(idx)
		var v Expr
		if fv.IsNil() {
			v = NullV{}
		} else {
			v = wrap(fv.Interface())
		}

		reprVals := map[string]string{}
		if tag, b := field.Tag.Lookup("faunarepr"); b {
			tagVals := strings.Split(tag, ",")
			for _, key := range tagVals {
				pair := strings.Split(key, "=")
				switch len(pair) {
				case 1:
					reprVals["fn"] = pair[0]
				case 2:
					reprVals[pair[0]] = pair[1]
				}
			}

			switch reprVals["fn"] {
			case "scopedfn":
				if (v != NullV{}) {
					name = "Scoped" + name
					if sbArgs.Len() > 0 {
						sbArgs.WriteString(", ")
					}
					sbArgs.WriteString(v.String())
				}
			case "optfn":
				if (v != NullV{}) {
					optFnName := field.Name
					if len(reprVals["name"]) != 0 {
						optFnName = reprVals["name"]
					}
					if sbOpt.Len()+sbArgs.Len() > 0 {
						sbOpt.WriteString(", ")
					}
					sbOpt.WriteString(optFnName)
					sbOpt.WriteString("(")
					if _, ok := reprVals["noargs"]; !ok {
						sbOpt.WriteString(v.String())
					}
					sbOpt.WriteString(")")
				}
			case "optarg":
				if (v != NullV{}) {
					if sbArgs.Len() > 0 {
						sbArgs.WriteString(", ")
					}
					sbArgs.WriteString(v.String())
				}
			case "varargs":
				if reflect.TypeOf(v).ConvertibleTo(reflect.TypeOf(unescapedArr{})) {
					nestedArgs := reflect.ValueOf(v).Interface().(unescapedArr)
					for _, nv := range nestedArgs {
						if sbArgs.Len() > 0 {
							sbArgs.WriteString(", ")
						}
						sbArgs.WriteString(nv.String())
					}
				} else {
					if sbArgs.Len()+sbOpt.Len() > 0 {
						sbArgs.WriteString(", ")
					}
					sbArgs.WriteString(v.String())
				}
			case "noargs":

			default:
				return "Unknown faunarepr: `" + reprVals["fn"] + "` in " + name
			}
		} else {
			if tag, b := field.Tag.Lookup("json"); b && strings.HasSuffix(tag, ",omitempty") {
				continue
			}
			if sbArgs.Len()+sbOpt.Len() > 0 {
				sbArgs.WriteString(", ")
			}
			sbArgs.WriteString(v.String())
		}
	}
	return name + "(" + sbArgs.String() + sbOpt.String() + ")"
}
