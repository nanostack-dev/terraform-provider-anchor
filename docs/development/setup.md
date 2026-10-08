# Local setup

Use the Go version in [go.mod](../../go.mod), currently 1.26.5. Terraform is needed to exercise provider configuration; ordinary Go unit tests do not require API credentials or a workspace checkout.

```sh
go mod download
go build -o /tmp/terraform-provider-anchor .
go vet ./...
go test ./...
```

For local Terraform development, use a developer-controlled CLI configuration with a `provider_installation.dev_overrides` entry for `nanostack-dev/anchor` pointing at the directory containing the locally built `terraform-provider-anchor` binary. [README](../../README.md) contains the complete example. With that override, use `terraform plan` directly rather than expecting `terraform init` to install the development binary. Review every plan before an apply against a development API.

The generated Anchor Go client is downloaded at its pinned version. Updating the API surface requires a matching published client and provider integration tests, not parent-directory imports. [Registry installation](../runbooks/deployment.md) describes released consumer usage.

## Agent harness setup

Codex, OpenCode and Grok Build discover the local `AGENTS.md` natively. Claude Code uses the project-local [.claude/settings.json](../../.claude/settings.json) SessionStart hook to read that guide from the Git checkout root, including sessions started in a nested directory. Approve project trust on first use and reload the session after adding or updating hooks. Personal overrides stay in ignored `.claude/settings.local.json`; required project guidance does not depend on the shared workspace or globally installed skills.
