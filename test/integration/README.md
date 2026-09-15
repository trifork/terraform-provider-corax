# Local integration harness

Applies a small but representative set of Corax resources against a **real**
Corax environment using a locally built provider binary. Unlike the acceptance
tests (`make testacc`), the state sticks around so you can inspect it, re-plan
for drift, and destroy when you are done.

## Setup

Put your credentials in the repo-root `.env` (gitignored, see `.env.example`):

```
CORAX_API_ENDPOINT=https://api.corax.app/
CORAX_API_KEY=<api-key>
```

## Usage

```shell
make integration-plan      # build provider + plan
make integration-apply     # create everything
make integration-plan      # should be empty -> no drift
make integration-destroy   # clean up (also clears the run_id)
```

Each of these rebuilds `bin/terraform-provider-corax` and writes a
`dev.tfrc` with a `dev_overrides` block, so **no `terraform init` is needed**
(and it would fail — the provider is not published). A `run_id` is generated
once into `run.auto.tfvars` so resource names are unique per run and stable
across plans; `integration-destroy` removes it.

## What it covers

| Resource | Exercises |
| --- | --- |
| `corax_project` | basic CRUD, computed counts |
| `corax_mcp_server` | `config` attribute map, `null` defaults, computed `slug` |
| `corax_chat_capability` | nested `config`, `data_retention = timed`, `blob_config`, dynamic `custom_parameters`, `mcp_server_ids` referencing another resource, `semantic_id` round-trip |
| `corax_completion_capability` | `output_type = "schema"` with `schema_def`, `variables` set, `data_retention = infinite` |
| `corax_extraction_capability` | nested `config`, `data_retention = timed`, `blob_config`, `custom_parameters`, `mcp_server_ids`, explicit `semantic_id`, plus a second instance with `semantic_id` omitted to cover the API-generated `Optional + Computed` round-trip |
| `corax_speech_to_text_capability` | opt-in via `enable_speech_to_text` (needs a default STT model) |
| `corax_api_key` | opt-in via `enable_api_key` (secret lands in state) |

Deliberately not covered: `corax_model_provider` and `corax_model_deployment`
(need real third-party provider credentials) and
`corax_capability_type_default_model` (mutates an environment-wide default).

Enable the opt-in resources with:

```shell
TF_VAR_enable_speech_to_text=true TF_VAR_enable_api_key=true make integration-apply
```

Exercise the in-place update path by bumping `revision`, which feeds into the
project description, both system prompts, and the chat temperature:

```shell
TF_VAR_revision=2 make integration-apply
TF_VAR_revision=2 make integration-plan   # should be empty again
```

## Known Corax API behaviours this harness encodes

- `config.content_tracing` **must** be `false` when `config.data_retention.type`
  is `"timed"`; the API forces it and the provider now rejects the combination
  at plan time.
- `variables` on a completion capability are accepted on write but never read
  back by the API, so the provider keeps the configured value.
- `schema_def` is a flat map of field name to property definition discriminated
  on `type`; an enum is `type = "enum"` with `enum = [...]`, **not**
  `type = "string"` plus `enum`. It is not a JSON Schema document.
- A create that fails after the API call leaves the `semantic_id` permanently
  reserved (`GET /v1/capabilities/<semantic_id>` then resolves to a UUID that
  404s), so a re-run needs a fresh `run_id`. That is a backend issue, not a
  provider one.

## Drift check

The most valuable signal here is the second `make integration-plan`: any
non-empty plan after a successful apply means the provider is not reading back
what it wrote (the classic cause being an API-computed field that is declared
`Optional` but not `Computed`).
