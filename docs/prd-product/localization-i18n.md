## 1. Feature Overview
- Name: Localization and Translation Delivery
- Summary: Backend-driven translation source of truth consumed by frontend `t(key)` usage.
- Target user: Multilingual end users.
- Business value: Enables language support with centralized key management.

## 2. Problem Statement
- Hardcoded UI text blocks international expansion and consistency.
- Without centralized i18n, frontend and backend messages diverge.

## 3. Current Behavior
- Translation files are JSON by locale under backend translations path.
- Backend resolves runtime locale from config/settings and serves translation JSON.
- Frontend loads translation payload once and resolves keys with optional interpolation.
- Missing key fallback returns original key string.

## 4. User Flows
### 4.1 Main Flow
1. App initializes i18n provider.
2. Provider fetches `/configuration/translation`.
3. UI calls `t('KEY', params)` and renders translated text.

### 4.2 Alternative Flows
1. User changes language in settings.
2. Backend reloads translations with new locale and serves new dictionary.

### 4.3 Error Scenarios
1. Invalid/unsafe locale resolves to default locale.
2. Missing translation key renders key fallback.
3. Translation load failure impacts text localization but app remains functional.

## 5. Functional Requirements
- The system must serve current locale translation dictionary.
- The user can switch language from settings.
- The system should safely sanitize locale path selection.

## 6. Business Rules
- Default locale fallback: `en-US`.
- Locale value must pass sanitization (no traversal chars, allowed charset).
- Interpolation uses `{{placeholder}}` replacement in frontend.

## 7. Data Model (Business View)
- Translation Dictionary: key -> localized string, per locale JSON file.

## 8. Interfaces
- User interfaces: all translated pages/components.
- API: `GET /configuration/translation`.

## 9. Dependencies
- Settings language selection.
- Backend i18n loader and translation file availability.

## 10. Limitations / Gaps
- No runtime key coverage diagnostics exposed to users.
- Fallback-to-key may leak internal key names in UI.

## 11. Opportunities
- Add translation completeness dashboard and locale health checks.
- Support pluralization/advanced formatting rules.

## 12. Acceptance Criteria
- Given locale is configured, when translation endpoint is requested, then locale dictionary is returned.
- Given unknown locale, when translation file path resolves, then default locale file is served.
- Given interpolation params, when `t` is called, then placeholders are replaced.

## 13. Assumptions
- Translation JSON files are deployed with backend artifacts.
