# Rollback

Pin the last verified provider version in the consumer's `required_providers`, remove any local development override if using Registry artifacts, and run [`terraform init -upgrade`](https://developer.hashicorp.com/terraform/cli/commands/init#plugin-installation) to resolve the revised constraint. Use an exact verified version for recovery; the flag reselects all providers and modules allowed by the configuration, so review every changed dependency selection. Review the new lock selection and `terraform plan` before applying. Preserve state securely and check that the older provider understands the current resource schema/state.

A provider binary rollback does not undo remote API mutations, destroyed resources or irreversible template archives. Restore or reconcile API resources through the owning service's documented lifecycle; do not rewrite Terraform state to hide an incompatible schema without a reviewed recovery procedure.

Revert or fix the source regression through a PR and publish a new corrective tag after validation. Preserve published version history and signed artifacts. Recovery evidence includes selected provider version, API target, reviewed plan and affected resource verification.
