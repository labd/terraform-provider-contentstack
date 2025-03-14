package provider

import (
	"fmt"
	"github.com/labd/contentstack-go-sdk/management"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func processResponse(resp any, input any) diag.Diagnostics {
	var diags diag.Diagnostics

	// Branches are only returned when the user has the contentstack plan
	// supporting this feature. We check if the interface has a field Branches,
	// if it has and it is empty we assume it's not part of the plan. Copy the
	// input value in that case so that terraform doesn't see vanishing
	// elements.
	t := reflect.TypeOf(resp)
	if _, ok := t.Elem().FieldByName("Branches"); ok {
		v := reflect.ValueOf(resp)
		branches := v.Elem().FieldByName("Branches")

		if branches.Len() == 0 {
			// Copy value from input to resp
			t := reflect.ValueOf(input).Elem().FieldByName("Branches")
			branches.Set(t)

			diags.AddAttributeWarning(
				tftypes.NewAttributePath().WithAttributeName("branches"),
				"Branches are not part of your plan.",
				"Branches are not part of your plan. Please contact support@contentstack.com to upgrade your plan.",
			)
		}
	}

	return diags
}

func copyHttpBasicPasswords(wd []management.WebhookDestination, data WebhookDestinationSlice) ([]management.WebhookDestination, error) {
	cd := make([]management.WebhookDestination, 0, len(wd))

	for _, d := range wd {
		planned := data.FindByTargetURLAndHttpBasicAuth(d.TargetURL, d.HttpBasicAuth)
		if planned == nil {
			return nil, fmt.Errorf("d %s not found in planned state", d.TargetURL)
		}

		d.HttpBasicPassword = planned.HttpBasicPassword.Value
		cd = append(cd, d)
	}
	return cd, nil
}
