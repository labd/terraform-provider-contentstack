# Contentstack Terraform Provider

The Terraform provider for [Contentstack](https://www.contentstack.com/) lets
you manage the configuration of a Contentstack stack as code. Content types,
global fields, environments, locales and webhooks are defined in Terraform, so
changes to a stack can be reviewed, versioned and applied the same way across
stacks and branches. It is built on the
[Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework)
and uses [contentstack-go-sdk](https://github.com/labd/contentstack-go-sdk) to
talk to the Contentstack Content Management API.

The full documentation is on the
[Terraform Registry](https://registry.terraform.io/providers/labd/contentstack/latest/docs).

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/install) 1.0 or newer,
  or [OpenTofu](https://opentofu.org/)
- [Go](https://go.dev/doc/install) 1.27 or newer, only to build the provider
  from source (see `go.mod`)

## Installation

Add the provider to your Terraform configuration and run `terraform init`:

```hcl
terraform {
  required_providers {
    contentstack = {
      source  = "labd/contentstack"
      version = "~> 0.4"
    }
  }
}
```

## Configuration

All provider attributes are optional in the schema, but `base_url` and
`api_key` are needed to reach a stack, plus one way to authenticate:

- `management_token`: a stack-level management token (recommended). It is sent
  as the `authorization` header.
- `auth_token`: a user-specific authtoken. It is only used when no
  `management_token` is set.

```hcl
provider "contentstack" {
  base_url         = "https://eu-api.contentstack.com/"
  api_key          = var.contentstack_api_key
  management_token = var.contentstack_management_token
  branch           = "main"

  # Optional: retries for 429 (rate limit) responses
  max_retries    = 3
  retry_wait_min = 1
  retry_wait_max = 30
}
```

| Attribute          | Description                                                                                                                              |
| ------------------ | ---------------------------------------------------------------------------------------------------------------------------------------- |
| `base_url`         | The Content Management API base URL for your region, see [Base URL](https://www.contentstack.com/docs/developers/apis/content-management-api/#base-url). |
| `api_key`          | The API key of the stack.                                                                                                                |
| `management_token` | Stack management token (sensitive).                                                                                                      |
| `auth_token`       | User authtoken (sensitive).                                                                                                              |
| `branch`           | The branch to manage resources in. Defaults to the main branch.                                                                          |
| `max_retries`      | Maximum retry attempts for 429 responses. Defaults to `3`.                                                                               |
| `retry_wait_min`   | Minimum wait in seconds between retries. Defaults to `1`.                                                                                |
| `retry_wait_max`   | Maximum wait in seconds between retries. Defaults to `30`.                                                                               |

The provider does not read any environment variables. To keep credentials out
of your configuration, pass them in through Terraform variables, for example
`TF_VAR_contentstack_management_token`.

## Usage

The provider has these resources:

| Resource                                                                                                          | Manages                    |
| ----------------------------------------------------------------------------------------------------------------- | -------------------------- |
| [`contentstack_content_type`](https://registry.terraform.io/providers/labd/contentstack/latest/docs/resources/content_type) | Content types              |
| [`contentstack_global_field`](https://registry.terraform.io/providers/labd/contentstack/latest/docs/resources/global_field) | Global fields              |
| [`contentstack_environment`](https://registry.terraform.io/providers/labd/contentstack/latest/docs/resources/environment)   | Publishing environments    |
| [`contentstack_locale`](https://registry.terraform.io/providers/labd/contentstack/latest/docs/resources/locale)             | Languages (locales)        |
| [`contentstack_webhook`](https://registry.terraform.io/providers/labd/contentstack/latest/docs/resources/webhook)           | Webhooks                   |

A content type with its schema given as JSON, a locale, an environment and a
webhook:

```hcl
resource "contentstack_content_type" "page" {
  title       = "Page"
  uid         = "page"
  description = "A simple page"

  schema = jsonencode([
    {
      display_name    = "Title"
      uid             = "title"
      data_type       = "text"
      mandatory       = true
      unique          = true
      multiple        = false
      non_localizable = false
    }
  ])
}

resource "contentstack_locale" "nl" {
  name            = "Nederlands"
  code            = "nl-nl"
  fallback_locale = "nl"
}

resource "contentstack_environment" "production" {
  name = "production"

  url {
    locale = "nl-nl"
    url    = "https://www.example.nl/"
  }
}

resource "contentstack_webhook" "publish" {
  name = "publish"

  destination {
    target_url          = "https://example.com/hooks/contentstack"
    http_basic_auth     = "user"
    http_basic_password = var.webhook_password

    custom_headers = [{
      header_name = "X-Custom"
      value       = "value"
    }]
  }

  channels        = ["assets.create"]
  branches        = ["main"]
  retry_policy    = "manual"
  disabled        = false
  concise_payload = true
}
```

All resources can be imported by their Contentstack UID:

```sh
terraform import contentstack_content_type.page page
```

More examples are in the [`examples/`](examples/) directory.

## Development

The repository uses [Task](https://taskfile.dev/) as a task runner:

```sh
task format     # go fmt ./...
task test       # go test -v ./...
task docs       # go generate: terraform fmt on examples/ and regenerate docs/
task coverage   # test coverage report
```

CI runs `go test -race -v ./...`, golangci-lint, govulncheck and
`go generate ./...`. It fails when the generated docs differ from what is
committed, so run `task docs` after changing a schema or an example. This needs
`terraform` on your `PATH`.

To try a local build, install it with `go install .` and point Terraform at it
with a `dev_overrides` block in `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "labd/contentstack" = "/path/to/your/go/bin"
  }
  direct {}
}
```

## Releasing

Releases are managed with [changie](https://github.com/miniscruff/changie):

1. Every change adds a changelog fragment with `changie new`. It is stored in
   `.changes/unreleased/`. Dependabot PRs get one automatically.
2. On every push to `main`, the
   [changie-release-action](https://github.com/labd/changie-release-action)
   opens or updates a release PR that batches the unreleased fragments and
   updates `CHANGELOG.md`.
3. Merging that PR tags the release and triggers the `release` workflow. That
   workflow runs [GoReleaser](https://goreleaser.com/) to build and sign the
   binaries and publish a GitHub release, which the Terraform Registry picks up.

## License

[MIT](LICENSE), copyright Lab Digital.
