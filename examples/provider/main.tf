terraform {
  required_providers {
    contentstack = {
      source = "labd/contentstack"
    }
  }
}

provider "contentstack" {
  base_url         = "https://api.contentstack.io"
  api_key          = "<api_key>"
  management_token = "<management_token>"
  branch           = "main"
}