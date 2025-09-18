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

  rate_limit  = 8.0
  rate_burst  = 5
  max_retries = 3
}

resource "contentstack_content_type" "example" {
  title = "Example Content Type"
  uid   = "example"

  schema = jsonencode([
    {
      display_name = "Title"
      uid          = "title"
      data_type    = "text"
      mandatory    = true
    }
  ])
}