# Testing

The [test workflow](../../.github/workflows/test.yml) runs these local checks:

```sh
go mod download
go build -v ./...
go vet ./...
go test -v ./...
```

Without `TF_ACC=1`, acceptance tests skip; local unit/surface/validator tests still run. [provider_test.go](../../internal/provider/provider_test.go) defines acceptance prerequisites and cleanup. Acceptance tests create real products and need a platform bearer token, not only a product API key.

After explicitly selecting a development target and supplying the token through an approved secret mechanism:

```sh
TF_ACC=1 go test -v ./...
```

Set `ANCHOR_BASE_URL` to the intended development API and `ANCHOR_TOKEN` before that command; never copy credentials into docs or command history. Report missing live acceptance evidence separately from passing unit checks. Resource/state/client upgrades need representative create/read/update/import/delete behavior and cleanup evidence.

Documentation-only changes need relative-link, command-provenance and `git diff --check` checks. Registry schema docs under `docs/resources` remain part of the review; there is no Markdown linter configured.

The test workflow runs on every pull request base so each layer created by
`gh stack` receives build, vet and unit checks. Push checks remain limited to
`main`. Live acceptance still requires the explicitly selected development API
and platform token above; an ordinary CI unit pass does not supply that evidence.
