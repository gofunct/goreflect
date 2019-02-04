package tf

import (
	"fmt"
	"github.com/zclconf/go-cty/cty/function"
	"reflect"
	"strings"

	"github.com/hashicorp/hcl2/hcl"
	"github.com/zclconf/go-cty/cty"
)

var specSchemaUnlabelled *hcl.BodySchema
var specSchemaLabelled *hcl.BodySchema

var specSchemaLabelledLabels = []string{"key"}

var typeType = cty.Capsule("type", reflect.TypeOf(cty.NilType))

var typeEvalCtx = &hcl.EvalContext{
	Variables: map[string]cty.Value{
		"string": wrapTypeType(cty.String),
		"bool":   wrapTypeType(cty.Bool),
		"number": wrapTypeType(cty.Number),
		"any":    wrapTypeType(cty.DynamicPseudoType),
	},
	Functions: map[string]function.Function{
		"list": function.New(&function.Spec{
			Params: []function.Parameter{
				{
					Name: "element_type",
					Type: typeType,
				},
			},
			Type: function.StaticReturnType(typeType),
			Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
				ety := unwrapTypeType(args[0])
				ty := cty.List(ety)
				return wrapTypeType(ty), nil
			},
		}),
		"set": function.New(&function.Spec{
			Params: []function.Parameter{
				{
					Name: "element_type",
					Type: typeType,
				},
			},
			Type: function.StaticReturnType(typeType),
			Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
				ety := unwrapTypeType(args[0])
				ty := cty.Set(ety)
				return wrapTypeType(ty), nil
			},
		}),
		"map": function.New(&function.Spec{
			Params: []function.Parameter{
				{
					Name: "element_type",
					Type: typeType,
				},
			},
			Type: function.StaticReturnType(typeType),
			Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
				ety := unwrapTypeType(args[0])
				ty := cty.Map(ety)
				return wrapTypeType(ty), nil
			},
		}),
		"tuple": function.New(&function.Spec{
			Params: []function.Parameter{
				{
					Name: "element_types",
					Type: cty.List(typeType),
				},
			},
			Type: function.StaticReturnType(typeType),
			Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
				etysVal := args[0]
				etys := make([]cty.Type, 0, etysVal.LengthInt())
				for it := etysVal.ElementIterator(); it.Next(); {
					_, wrapEty := it.Element()
					etys = append(etys, unwrapTypeType(wrapEty))
				}
				ty := cty.Tuple(etys)
				return wrapTypeType(ty), nil
			},
		}),
		"object": function.New(&function.Spec{
			Params: []function.Parameter{
				{
					Name: "attribute_types",
					Type: cty.Map(typeType),
				},
			},
			Type: function.StaticReturnType(typeType),
			Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
				atysVal := args[0]
				atys := make(map[string]cty.Type)
				for it := atysVal.ElementIterator(); it.Next(); {
					nameVal, wrapAty := it.Element()
					name := nameVal.AsString()
					atys[name] = unwrapTypeType(wrapAty)
				}
				ty := cty.Object(atys)
				return wrapTypeType(ty), nil
			},
		}),
	},
}

func ParseVarsArg(src string, argIdx int) (map[string]cty.Value, hcl.Diagnostics) {
	fakeFn := fmt.Sprintf("<vars argument %d>", argIdx)
	f, diags := parser.ParseJSON([]byte(src), fakeFn)
	if f == nil {
		return nil, diags
	}
	vals, valsDiags := ParseVarsBody(f.Body)
	diags = append(diags, valsDiags...)
	return vals, diags
}

func ParseVarsFile(filename string) (map[string]cty.Value, hcl.Diagnostics) {
	var f *hcl.File
	var diags hcl.Diagnostics

	if strings.HasSuffix(filename, ".json") {
		f, diags = parser.ParseJSONFile(filename)
	} else {
		f, diags = parser.ParseHCLFile(filename)
	}

	if f == nil {
		return nil, diags
	}

	vals, valsDiags := ParseVarsBody(f.Body)
	diags = append(diags, valsDiags...)
	return vals, diags

}

func ParseVarsBody(body hcl.Body) (map[string]cty.Value, hcl.Diagnostics) {
	attrs, diags := body.JustAttributes()
	if attrs == nil {
		return nil, diags
	}

	vals := make(map[string]cty.Value, len(attrs))
	for name, attr := range attrs {
		val, valDiags := attr.Expr.Value(nil)
		diags = append(diags, valDiags...)
		vals[name] = val
	}
	return vals, diags
}
