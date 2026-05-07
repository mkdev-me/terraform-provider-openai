# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed
- Admin-API calls are now paced by a token-bucket rate limiter (default 6
  RPM with a burst of 4) in addition to the v2.2.6 concurrency semaphore.
  Empirical testing on 2026-05-07 found the admin API throttles per-endpoint
  at ~7-10 RPM (no `x-ratelimit-*` or `Retry-After` headers exposed): seven
  sequential `GET /v1/projects/{id}/roles` calls succeed, the next seven in
  the same minute return 429, and the bucket refills after ~60s. The
  v2.2.6 concurrency cap was therefore insufficient — even one-in-flight
  sequential calls 429 once the per-minute bucket is exhausted. A real
  `terraform plan` against 14 distinct projects now completes without 429s
  (vs. failing under v2.2.6). Override defaults via `OPENAI_ADMIN_MAX_RPM`
  (clamped `[1, 600]`) and `OPENAI_ADMIN_BURST` (clamped `[1, 100]`).

## [2.2.6]

### Fixed
- All admin-API calls now pass through a per-process counting semaphore
  (default 3 in-flight, override via `OPENAI_ADMIN_MAX_CONCURRENT`,
  clamped to `[1, 64]`). Terraform's parallelism (default 10) was bursting
  requests faster than the per-request retry+jitter (v2.2.3) could absorb;
  the retries themselves stacked under load and kept colliding with the
  rate-limit window. Serialising at the provider level — independent of
  Terraform's parallelism flag — is necessary on top of retry+jitter. (See
  v2.2.7 for the additional rate-limit fix once empirical testing showed
  concurrency capping alone was still insufficient against the admin
  API's per-minute ceiling.)
- `openai_project_group` Create no longer issues a paginated list of all
  groups in the project after a successful POST. The POST response itself
  is the project-group object, so we parse it directly. The list-and-find
  fallback now runs only on the idempotent "already exists in project"
  path. During a parallel apply attaching N groups to one project, this
  reduces 2N admin-API calls (POST + paginated list per group) to N (POST
  only), which keeps the admin rate limit from 429'ing the apply mid-flight.

## [2.2.4]

### Fixed
- The `openai_project_role` data source (singular and plural) and the
  `openai_group` data source (singular and plural) now share their admin-API
  list calls across invocations via per-process caches, so a `terraform
  plan` that declares N role lookups against M projects and K group lookups
  resolves to M + 1 admin-API list calls instead of N + K. Roles are cached
  per project ID; groups are cached once for the org. Without this, retry
  alone (added in v2.2.3) was insufficient — concurrent paginated lookups
  for the same data still bursted past the admin rate limit and 429'd
  repeatedly even with jittered backoff.

## [2.2.3]

### Fixed
- All admin-API HTTP calls in the `openai_project_group` resource (Create/
  Read/Update/Delete), the `openai_project_user` resource (Create/Read/
  Update/Delete), and the `openai_group` data source now retry on `429 Too
  Many Requests` and transient `5xx` responses, using the same backoff
  helper introduced in v2.2.1. Previously only the `openai_project_role`
  data source (v2.2.2) and the v0→v1 state upgrader (v2.2.1) were covered,
  leaving plans that read groups and applies that create/update/delete
  project memberships exposed: a `terraform apply` attaching ~14 groups to
  a single project consistently failed with `API error listing project
  groups: 429`.
- Backoff schedule now applies full jitter (actual sleep is uniformly
  random in `[base/2, base]` for base values 1, 2, 4, 8, 16, 30s). Without
  jitter, N concurrent requests that all 429 at the same moment retry on
  the same instants and keep colliding; jitter spreads them out and lets
  the rate-limit window drain.

## [2.2.2]

### Fixed
- The `openai_project_role` data source (singular and plural) now retries on
  `429 Too Many Requests` and transient `5xx` responses from the admin API,
  using the same exponential backoff helper that the state upgrader uses
  (introduced in v2.2.1). Without retry, a `terraform plan` declaring many
  `data "openai_project_role"` lookups against the same admin API can burst
  past the rate limit and fail the entire plan with `Status: 429 Too Many
  Requests`.

## [2.2.1]

### Fixed
- The `openai_project_user` state upgrader (added in v2.2.0) now retries on
  `429 Too Many Requests` and transient `5xx` responses from the admin API
  using exponential backoff (1s, 2s, 4s, 8s, 16s, 30s). The admin
  `GET /v1/projects/{id}/roles` endpoint is rate-limited; without retry, a
  state upgrade migrating many resources can burst past the limit and fail
  the entire `terraform plan`. Honours the `Retry-After` header when present.

## [2.2.0]

### Added
- State upgrader for `openai_project_user` to migrate state stored under the
  pre-v2.1.0 schema (string `role`) into the current schema (set `role_ids`).
  Users upgrading from v2.0.0 no longer need manual `terraform state rm` /
  `import`: the upgrader resolves the prior role name to its role ID via the
  admin API and writes it into `role_ids` transparently on the next plan.
  Lookups are cached per project inside the provider process so a plan over
  many resources in the same project performs a single role-list call rather
  than one per resource.

### Migration notes
- The upgrader requires an **admin API key** (the same one already required
  to manage project users in normal operation). Operators whose existing v0
  state was created without admin credentials configured must provision an
  admin key before running `terraform plan` against the new provider,
  otherwise the upgrade fails with
  `admin API key is required to resolve project role`.
- The upgrader fails closed when the prior role name no longer exists in the
  project (renamed/deleted role). This is intentional: silently writing
  `role_ids: []` would revoke access on the next apply. If you hit this,
  investigate the affected project before retrying.
- `openai_project_group` was introduced in v2.1.0 and has no prior released
  schema, so no migration is needed for it.

## [1.1.0] - 2025-06-27
### Added
- Timeout configuration support for provider operations (#21)
- Update and delete functionality for Organization Users resource (#19)

### Changed
- Simplified import scripts for better maintainability (#18)
- Updated example scripts for clarity (#17)

### Fixed
- Docs updated to render examples on Terraform Registry correctly

## [1.0.3] - 2025-06-23
### Fixed
- Docs updated to render examples on Terraform Registry correctly

## [1.0.2] - 2025-06-21

### Fixed
- Project archival / deletion

## [1.0.1] - 2025-06-21

### Fixed
- Fixed pagination issue when fetching all projects - now retrieves all pages instead of just the first page
- Cleaned up hardcoded version references in example files
- Removed unnecessary API key mentions from examples
- Improved code organization by reordering imports

## [1.0.0] - 2025-06-20

### Added
- Initial release of the Terraform Provider for OpenAI
- Provider configuration with support for API keys and organization ID
- Resource: `openai_chat_completion` - Manage chat completions
- Resource: `openai_embedding` - Create embeddings
- Resource: `openai_file` - Manage files for fine-tuning and assistants
- Resource: `openai_fine_tuning_job` - Create and manage fine-tuning jobs
- Resource: `openai_image` - Generate and edit images
- Resource: `openai_audio_transcription` - Transcribe audio files
- Resource: `openai_audio_translation` - Translate audio files
- Resource: `openai_audio_speech` - Generate speech from text
- Resource: `openai_assistant` - Manage AI assistants
- Resource: `openai_assistant_file` - Attach files to assistants
- Resource: `openai_thread` - Manage conversation threads
- Resource: `openai_message` - Create messages in threads
- Resource: `openai_run` - Execute assistant runs
- Resource: `openai_vector_store` - Manage vector stores
- Resource: `openai_vector_store_file` - Manage files in vector stores
- Resource: `openai_vector_store_file_batch` - Batch operations for vector store files
- Resource: `openai_organization_invite` - Manage organization invitations
- Resource: `openai_organization_user` - Manage organization users
- Resource: `openai_project` - Manage projects
- Resource: `openai_project_rate_limit` - Configure project rate limits
- Resource: `openai_project_service_account` - Manage project service accounts
- Resource: `openai_project_user` - Manage project users
- Data Source: `openai_file` - Read file information
- Data Source: `openai_fine_tuning_job` - Read fine-tuning job information
- Data Source: `openai_model` - Get model information
- Data Source: `openai_models` - List available models
- Data Source: `openai_organization_audit_logs` - Read organization audit logs
- Data Source: `openai_organization_invites` - List organization invitations
- Data Source: `openai_organization_project` - Read project information
- Data Source: `openai_organization_projects` - List organization projects
- Data Source: `openai_organization_users` - List organization users
- Data Source: `openai_project_api_key` - Read project API key information
- Data Source: `openai_project_api_keys` - List project API keys
- Data Source: `openai_project_rate_limits` - List project rate limits
- Data Source: `openai_project_service_account` - Read service account information
- Data Source: `openai_project_service_accounts` - List project service accounts
- Data Source: `openai_project_user` - Read project user information
- Data Source: `openai_project_users` - List project users
- Comprehensive documentation for all resources and data sources
- Example configurations for common use cases
- Reusable Terraform modules for common patterns

[Unreleased]: https://github.com/mkdev-me/terraform-provider-openai/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/mkdev-me/terraform-provider-openai/releases/tag/v0.1.0
