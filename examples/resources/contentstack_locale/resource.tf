
terraform {
  required_providers {
    contentstack = {
      source = "labd/contentstack"
    }
  }
}

provider "contentstack" {
  base_url         = "https://eu-api.contentstack.com/"
  api_key          = "<api_key>"
  management_token = "<token>"
}


resource "contentstack_locale" "nl" {
  name            = "Nederlands"
  code            = "nl-nl"
  fallback_locale = "nl"
}
