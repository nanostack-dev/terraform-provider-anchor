# Domain context

| Term | Meaning |
| --- | --- |
| Product | Anchor application boundary identified by a public KSUID. |
| Product role | Named role owned by a product. |
| Product permission | Product resource/permission definition managed by the API. |
| License schema | Product licensing structure managed as configuration. |
| License template | Reusable license configuration; may be archived while preserving references/history. |
| Organization license | Runtime copy of a template with customer-specific adjustments; Terraform does not manage or expose it. |
| Platform token | Bearer credential used for platform operations, including acceptance tests which create products. |
| Product API key | Product-scoped credential paired with a `product_id`. |
| Drift | Difference between Terraform configuration/state and the current API state, including admin UI edits. |
| Archive | Irreversible withdrawal of a license template, preserving its resource/history. |

[Resource documentation](docs/README.md) owns exact fields, import forms and lifecycle behavior. [Architecture](docs/technical/architecture.md) records ownership boundaries.
