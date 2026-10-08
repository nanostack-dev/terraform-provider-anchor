# Anchor Terraform Provider

Terraform Plugin Framework provider for Anchor products, roles, permissions, license schemas and templates. Build/test procedures and contracts are local to this standalone checkout.

- Before implementation or delivery, read [agent workflow](docs/development/agent-workflow.md) and [testing](docs/development/testing.md).
- Before resources, state, import or authentication changes, read [architecture](docs/technical/architecture.md), [domain context](CONTEXT.md) and the relevant [resource docs](docs/README.md). Organization licenses are runtime data and remain outside Terraform ownership.
- Before setup or recurring failures, read [setup](docs/development/setup.md) or [troubleshooting](docs/development/troubleshooting.md). Acceptance tests create real resources and need an explicitly selected development API and platform token.
- Before publication, read [deployment](docs/runbooks/deployment.md); for version/state recovery, read [rollback](docs/runbooks/rollback.md).
- The Anchor Go client is a pinned generated dependency. Coordinate API/client/resource changes with Anchor, preserving tenant scope and permission checks; validate local provider behavior plus matching API integration evidence.
- Keep required guidance local and update owning docs in the same PR as behavior or verified lessons.
