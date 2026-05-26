package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"testing"

	"github.com/labd/contentstack-go-sdk/management"
	"github.com/stretchr/testify/assert"
)

type MockResponse struct {
	Branches []string
}

type MockInput struct {
	Branches []string
}

func TestProcessResponseWithBranches(t *testing.T) {
	resp := &MockResponse{Branches: []string{}}
	input := &MockInput{Branches: []string{"branch1"}}

	diags := processResponse(resp, input)

	assert.Len(t, diags, 1)
	assert.Equal(t, "branch1", resp.Branches[0])
}

func TestProcessResponseWithoutBranchesField(t *testing.T) {
	resp := &struct{ NoBranchesField string }{}
	input := &MockInput{Branches: []string{"branch1"}}

	diags := processResponse(resp, input)

	assert.Len(t, diags, 0)
}

func TestCopyHttpBasicPasswordsSuccess(t *testing.T) {
	wd := []management.WebhookDestination{
		{TargetURL: "http://example.com", HttpBasicAuth: "auth1"},
	}
	data := WebhookDestinationSlice{
		{TargetURL: types.StringValue("http://example.com"), HttpBasicAuth: types.StringValue("auth1"), HttpBasicPassword: types.StringValue("password1")},
	}

	result, err := copyHttpBasicPasswords(wd, data)

	assert.NoError(t, err)
	assert.Equal(t, "password1", result[0].HttpBasicPassword)
}

func TestCopyHttpBasicPasswordsNotFound(t *testing.T) {
	wd := []management.WebhookDestination{
		{TargetURL: "http://example.com", HttpBasicAuth: "auth1"},
	}
	data := WebhookDestinationSlice{}

	result, err := copyHttpBasicPasswords(wd, data)

	assert.Error(t, err)
	assert.Nil(t, result)
}
