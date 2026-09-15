# Local Terraform integration harness + provider fixes

## Done

- [x] `test/integration/` harness (`main.tf`, `variables.tf`, `outputs.tf`, README)
      applying against a real Corax environment via a locally built binary.
- [x] `GNUmakefile` targets `integration-{setup,plan,apply,destroy}` using
      `dev_overrides` (no `terraform init` needed) and a stable per-run `run_id`.
- [x] Fix: `custom_parameters` rejected HCL number and list literals
      (`types.Number` / `types.Tuple` were unhandled).
- [x] Fix: `config.mcp_server_ids` mapped the API's `[]` to an empty list,
      conflicting with a null config value.
- [x] Fix: `content_tracing = true` with timed data retention now errors at plan
      time instead of failing at apply (the API forces it to false).
- [x] Fix: completion `variables` were nulled from the create response even
      though the API never echoes them back.
- [x] New: `schema_def` plan-time validation, so a wrong property shape (e.g.
      `type = "string"` plus `enum`) errors clearly instead of being silently
      dropped by the API and surfacing as "bug in the provider".

## Review

Verified against `https://api.corax.app/`: `apply` -> `plan` (no drift) ->
`apply -var revision=2` (in-place update) -> `plan` (no drift) -> `destroy`,
for all six covered resources including the two opt-in ones. `go test ./...`
and `golangci-lint run` are green; `make generate` refreshed the docs.

## Open, not provider-side

A capability create that errors after the API call burns its `semantic_id`
permanently: the semantic ID keeps resolving to a UUID that returns 404, and
the ID cannot be reused or deleted. Worth reporting to the backend team.
