# Troubleshooting

| Symptom | Cause and repair | Verification |
| --- | --- | --- |
| Provider reports missing product ID | Product API key authentication requires product scope. Supply `product_id` or `ANCHOR_PRODUCT_ID`. | Configuration succeeds and scoped reads target the intended product. |
| Acceptance tests skip | `TF_ACC` is unset. Run the explicit acceptance procedure only against an approved development target. | Acceptance cases execute and cleanup completes. |
| Acceptance setup rejects credentials | Tests create their own product and require `ANCHOR_TOKEN`. Supply a platform bearer token for the selected API. | Test setup and resource lifecycle complete. |
| Destroying a license template is rejected | An organization license references it. Preserve the customer license; evaluate archive withdrawal using the resource documentation. | The chosen lifecycle operation and subsequent read return the expected state. |
| Terraform plans to undo admin edits | The provider has no ownership marker; UI edits are ordinary drift. Review configuration and reconcile the intended values. | A reviewed plan reflects the intended ownership and changes. |
| Local binary was not selected | Development overrides use the binary directory and bypass Registry initialization. Follow the README override configuration. | Terraform's plan uses the local provider. |

Archiving is irreversible. A binary rollback cannot undo an archive or restore destroyed remote resources. Keep state, tokens and customer data out of shared troubleshooting examples.
