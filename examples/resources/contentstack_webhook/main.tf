
terraform {
  required_providers {
    contentstack = {
      source  = "labd/contentstack"
      version = "0.1.0"
    }
  }
}

provider "contentstack" {
  base_url         = "https://eu-api.contentstack.com/"
  api_key          = "<api_key>"
  management_token = "<token>"
}


resource "contentstack_webhook" "mywebhook" {
  name = "test"

  destination {
    target_url          = "http://example.com"
    http_basic_auth     = "user"
    http_basic_password = "password"

    custom_headers = [{
      header_name = "Custom"
      value       = "testing"
    }]
  }

  channels = ["assets.create"]
  branches = ["main"]

  retry_policy    = "manual"
  disabled        = false
  concise_payload = true
}
