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

  # Optional: retry configuration for 429 responses
  max_retries    = 3
  retry_wait_min = 1
  retry_wait_max = 30
}
