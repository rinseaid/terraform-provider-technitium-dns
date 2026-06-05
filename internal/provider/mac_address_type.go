package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type macAddressType struct {
	basetypes.StringType
}

func (t macAddressType) Equal(o attr.Type) bool {
	other, ok := o.(macAddressType)
	if !ok {
		return false
	}
	return t.StringType.Equal(other.StringType)
}

func (t macAddressType) String() string {
	return "macAddressType"
}

func (t macAddressType) ValueFromString(_ context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	return macAddressValue{StringValue: in}, nil
}

func (t macAddressType) ValueType(_ context.Context) attr.Value {
	return macAddressValue{}
}

type macAddressValue struct {
	basetypes.StringValue
}

func (v macAddressValue) Type(_ context.Context) attr.Type {
	return macAddressType{}
}

func (v macAddressValue) Equal(o attr.Value) bool {
	other, ok := o.(macAddressValue)
	if !ok {
		return false
	}
	return v.StringValue.Equal(other.StringValue)
}

func (v macAddressValue) StringSemanticEquals(_ context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	newValue, ok := newValuable.(macAddressValue)
	if !ok {
		diags.AddError(
			"Semantic Equality Check Error",
			fmt.Sprintf("Expected macAddressValue, got: %T", newValuable),
		)
		return false, diags
	}

	return normalizeMAC(v.ValueString()) == normalizeMAC(newValue.ValueString()), diags
}
