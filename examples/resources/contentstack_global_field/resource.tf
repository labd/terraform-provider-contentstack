
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


resource "contentstack_global_field" "my_field" {
  title              = "test something"
  uid                = "foobar"
  description        = "someting"
  maintain_revisions = true

  schema = jsonencode(jsondecode(<<JSON
    [
      {
        "display_name": "Name",
        "uid": "name",
        "multiple": false,
        "non_localizable": false,
        "unique": false,
        "mandatory": false,
        "data_type": "text"
      },
      {
        "data_type": "text",
        "display_name": "Rich text editor",
        "uid": "description",
        "field_metadata": {
          "allow_rich_text": true,
          "description": "foobar",
          "multiline": false,
          "rich_text_type": "advanced",
          "options": [],
          "version": 3
        },
        "multiple": false,
        "non_localizable": false,
        "mandatory": false,
        "unique": false
      }
    ]
JSON
  ))
}
