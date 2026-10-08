# Publication and consumer installation

The provider publishes signed platform archives; it does not deploy the Anchor API. [release.yml](../../.github/workflows/release.yml) runs on pushed `v*` tags. Unlike the runner/kit repositories, a merge to `main` alone does not publish a version.

Before an approved release, pass build/vet/unit checks and the affected acceptance coverage, update [CHANGELOG](../../CHANGELOG.md) and review Registry schema examples. Choose a new semantic tag for the verified commit; the workflow uses [GoReleaser](../../.goreleaser.yml), imports the configured signing key, signs SHA-256 checksum files and attaches the Terraform Registry manifest. Signing secrets and Registry registration remain operator-owned prerequisites; [README](../../README.md) documents their setup.

Verify the Release workflow, platform zip archives, manifest, checksums and detached signature. After Registry ingestion, consumers select `nanostack-dev/anchor` and an appropriate version constraint in `required_providers`, then run `terraform init` and review `terraform plan`. [examples/provider/provider.tf](../../examples/provider/provider.tf) and [Registry guide](../index.md) own configuration examples.

Verify installed provider selection and a scoped representative resource read/plan against the intended API. A GitHub release alone does not prove Registry ingestion or live API compatibility.
