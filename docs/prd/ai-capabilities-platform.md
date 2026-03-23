## 1. Feature Overview

* Name: AI Capabilities Platform
* Summary: Shared AI service/router used by multiple features (search expansion, analytics insight, video description enrichment, image classification fallback).
* Purpose: Provide provider-agnostic AI execution with task-based routing and retries.
* Business value: Enables smarter product experiences without hard-coupling domains to one LLM provider.

## 2. Current Implementation

* How it works today: AI service is initialized when API keys exist; router maps task types to providers (OpenAI/Anthropic) with retry/backoff behavior.
* Main flows:
  * Domain feature builds task prompt and executes AI request via shared service.
  * Router selects provider by task and configured fallback policy.
  * Provider executes external API call and returns normalized response content.
* Entry points (routes, handlers, jobs):
  * No direct public endpoint; called internally by:
  * Search service (`TaskExtraction`)
  * Analytics service (insight generation)
  * Video service (catalog descriptions)
  * File image classification fallback
* Key files involved (list with paths):
  * `backend/pkg/ai/service.go`
  * `backend/pkg/ai/router.go`
  * `backend/pkg/ai/types.go`
  * `backend/pkg/ai/config.go`
  * `backend/pkg/ai/providers/openai/provider.go`
  * `backend/pkg/ai/providers/anthropic/provider.go`
  * `backend/pkg/ai/prompts/prompts.go`

## 3. Architecture & Design

* Layers involved (frontend/backend):
  * Backend shared package consumed by domain services.
* Data flow (step-by-step):
  * On app boot, config loads API keys/provider settings.
  * Service receives `ai.Request` (task type, prompts, tokens, temperature).
  * Router picks provider and optional fallback.
  * Provider client performs HTTP call and maps to `ai.Response`.
  * Caller parses content into feature-specific schema.
* External integrations:
  * OpenAI API
  * Anthropic API
* State management (if applicable):
  * Stateless request/response interactions with retry logic.

## 4. Data Model

* Entities involved:
  * AI request/response, provider config, task type mapping.
* Database tables / schemas:
  * None in this shared package.
* Relationships:
  * Task types map to provider strategies and prompt templates.
* Important fields:
  * `TaskType`, `SystemPrompt`, `Prompt`, `MaxTokens`, `Temperature`, provider ids.

## 5. Business Rules

* Explicit rules implemented in code:
  * AI is disabled if no provider API keys are configured.
  * Provider routing/fallback is task-aware.
  * Retry applies for retryable failures.
* Edge cases handled:
  * AI call failures should degrade gracefully in caller features.
* Validation logic:
  * Config loading and provider availability checks at startup.

## 6. User Flows

* Normal flow:
  * User triggers AI-enabled feature indirectly (search/analytics/video/images) and receives enriched result.
* Error flow:
  * AI provider timeout/error is logged and feature falls back to non-AI behavior where implemented.
* Edge cases:
  * Malformed AI output (e.g., JSON parsing) is tolerated by callers with fallback behavior.

## 7. API / Interfaces

* Endpoints:
  * No direct HTTP endpoints.
* Input/output:
  * Internal interface input: `ai.Request`; output: `ai.Response`.
* Contracts:
  * Callers must handle plain-text response and schema parsing.
* Internal interfaces:
  * `ai.ServiceInterface`
  * provider interfaces in router/service.

## 8. Problems & Limitations

* Technical debt:
  * Prompt/output schema governance is distributed across features.
* Bugs or inconsistencies:
  * Different features parse AI responses differently, increasing inconsistency risk.
* Performance issues:
  * AI requests can add latency when executed inline on request path.
* Missing validations:
  * No universal structured-output validator shared across features.

## 9. Security Concerns ⚠️

* Any suspicious behavior:
  * No obfuscated code found.
* External code execution:
  * External API invocation to third-party AI providers.
* Unsafe patterns:
  * Potential prompt-injection/data-exfiltration risk when untrusted content is sent to AI.
* Injection risks:
  * Model output is untrusted text and must be validated/sanitized by consuming modules.
* Hardcoded secrets:
  * No hardcoded API keys observed; keys are env-config driven.
* Unsafe file/system access:
  * None direct in AI package.

## 10. Improvement Opportunities

* Refactors:
  * Create centralized structured-output parsing/validation utilities.
* Architecture improvements:
  * Add policy layer for allowed prompts/data classes per task.
* Scalability:
  * Add queueing/caching for expensive AI calls.
* UX improvements:
  * Surface explainability/fallback notices when AI enhancement is unavailable.

## 11. Acceptance Criteria

* Functional:
  * AI-enabled features execute via shared service when configured.
  * Service cleanly degrades when unavailable.
* Technical:
  * Task routing and fallback provider logic works per configuration.
  * Retry behavior handles transient provider failures.
* Edge cases:
  * Invalid provider responses do not crash caller features.

## 12. Open Questions

* Unknown behaviors:
  * Preferred data retention/privacy policy for prompts and responses.
* Missing clarity in code:
  * No explicit cost-governance or rate-limit strategy across tasks.
* Assumptions made:
  * AI features are optional enhancements, not hard dependencies for core functionality.
