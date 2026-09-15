# Copyright (c) Trifork

variable "name_prefix" {
  description = "Prefix applied to every resource created by this harness."
  type        = string
  default     = "tf-acc"
}

variable "run_id" {
  description = "Unique suffix for this run. `make integration-apply` generates one into run.auto.tfvars."
  type        = string
}

variable "revision" {
  description = "Bump this (e.g. -var revision=2) to exercise the in-place update path."
  type        = number
  default     = 1
}

variable "enable_speech_to_text" {
  description = "Create a speech-to-text capability. Requires a default speech_to_text model on the target environment."
  type        = bool
  default     = false
}

variable "enable_api_key" {
  description = "Create an API key. Disabled by default because the secret lands in state."
  type        = bool
  default     = false
}
