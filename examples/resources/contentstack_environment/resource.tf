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

resource "contentstack_environment" "production" {
  name = "production"

  url {
    locale = "nl"
    url    = "https://www.labdigital.nl/"
  }

  url {
    locale = "gb"
    url    = "http://example.com"
  }
}
