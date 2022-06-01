
terraform {
  required_providers {
    contentstack = {
      source = "labd/contentstack"
    }
  }
}

provider "contentstack" {
  base_url         = "https://eu-api.contentstack.com/"
  api_key          = "foobar"
  management_token = "foobar"
}


resource "contentstack_webhook" "mywebhook" {
  name = "test"

  destination {
    target_url          = "http://example.com"
    http_basic_auth     = "basic"
    http_basic_password = "test"

    custom_header {
      header_name = "Custom"
      value       = "testing"
    }
  }

  channels = ["assets.create"]
  branches = ["main"]

  retry_policy    = "manual"
  disabled        = false
  concise_payload = true
}
