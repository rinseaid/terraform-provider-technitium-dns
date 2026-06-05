package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
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

func (t macAddressType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.StringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}
	stringValue, ok := attrValue.(basetypes.StringValue)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}
	stringValuable, diags := t.ValueFromString(ctx, stringValue)
	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting StringValue to StringValuable: %v", diags)
	}
	return stringValuable, nil
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

type requiresReplaceIfMACChanged struct{}

func (m requiresReplaceIfMACChanged) Description(_ context.Context) string {
	return "Requires replacement if the normalized MAC address changes."
}

func (m requiresReplaceIfMACChanged) MarkdownDescription(_ context.Context) string {
	return "Requires replacement if the normalized MAC address changes."
}

func (m requiresReplaceIfMACChanged) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() || req.PlanValue.IsNull() {
		resp.RequiresReplace = true
		return
	}
	if normalizeMAC(req.StateValue.ValueString()) != normalizeMAC(req.PlanValue.ValueString()) {
		resp.RequiresReplace = true
	}
}
