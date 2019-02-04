package tf

import (
	"fmt"
	"reflect"

	"github.com/hashicorp/hcl2/hcl"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

func evalTypeExpr(expr hcl.Expression) (cty.Type, hcl.Diagnostics) {
	result, diags := expr.Value(typeEvalCtx)
	if result.IsNull() {
		return cty.DynamicPseudoType, diags
	}
	if !result.Type().Equals(typeType) {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid type expression",
			Detail:   fmt.Sprintf("A type is required, not %s.", result.Type().FriendlyName()),
		})
		return cty.DynamicPseudoType, diags
	}

	return unwrapTypeType(result), diags
}

func wrapTypeType(ty cty.Type) cty.Value {
	return cty.CapsuleVal(typeType, &ty)
}

func unwrapTypeType(val cty.Value) cty.Type {
	return *(val.EncapsulatedValue().(*cty.Type))
}
