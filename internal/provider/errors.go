package provider

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/labd/contentstack-go-sdk/management"
)

func processRemoteError(e error) diag.Diagnostics {
	var diags diag.Diagnostics

	switch err := e.(type) {

	case *management.ErrorMessage:
		errors := []string{}
		for fieldName, fieldErrors := range err.Errors {
			for _, msg := range fieldErrors {
				errors = append(errors, fmt.Sprintf(" %s - %s", fieldName, msg))
			}
		}
		diags.AddError(err.ErrorMessage, strings.Join(errors, "\n"))

	default:
		diags.AddError(e.Error(), e.Error())
	}
	return diags

}

func IsNotFoundError(e error) bool {
	if e == nil {
		return false
	}

	if err, ok := e.(*management.ErrorMessage); ok {
		return err.ErrorCode == 404
	}

	return false
}
