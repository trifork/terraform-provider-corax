# Lessons

## Unit tests must use the value types Terraform actually produces

`custom_parameters` (a `DynamicAttribute`) had thorough unit tests built from
`types.Int64Value` / `types.Float64Value` / `types.ListValueMust`, and they all
passed -- but Terraform hands a dynamic attribute a bare HCL number as
`types.Number` and a bare HCL list as `types.Tuple`, neither of which the
converter handled. Every real `apply` with a numeric custom parameter failed.

**Rule:** when testing a conversion that sits on a Terraform boundary, build the
input from the types the wire protocol produces (`Number`, `Tuple`, `Set`), not
the ergonomic Go-side types. If a unit test cannot tell you that, get a real
`apply` to tell you -- that is what `test/integration` is for.

## "Optional + Computed" does not let a provider override a configured value

`content_tracing` was `Optional + Computed` on the assumption the API could
override it. Terraform still requires the applied value to equal the configured
value, so the API forcing `false` surfaced as "Provider produced inconsistent
result after apply". A cross-field validator that rejects the combination at
plan time is the correct fix, not a plan modifier.
