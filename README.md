# Contentstack Terraform Provider

The Terraform provider for [Contentstack](https://www.contentstack.com/) allows 
you to configure your Contenstack stack with infrastructure-as-code principles.

## Usage

The full documentation is available via https://registry.terraform.io/providers/labd/contentstack/latest/docs


Add the following to your terraform project:

```hcl
terraform {
  required_providers {
    contentstack = {
      source = "labd/contentstack"
    }
  }
}
```

## Authors

This project is developed by [Lab Digital](https://www.labdigital.nl). We
welcome additional contributors. Please see our
[GitHub repository](https://github.com/labd/terraform-provider-contentstack)
for more information.
