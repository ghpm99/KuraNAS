# Mobile KuraNAS

Aplicativo Android nativo do KuraNAS com foco em compatibilidade com Android 4.1.2 (API 16).

## Restrições Obrigatórias

- Dispositivo alvo: Samsung Galaxy Tab 2 7.0 (GT-P3110), 1024x600.
- Stack obrigatória: `Java + XML Views + AppCompat`.
- Não usar Kotlin.
- Não usar Jetpack Compose.
- Toda decisão de API/lib deve ser compatível com `minSdk 16`.

## Estrutura

```text
mobile/
├── app/
│   ├── src/main/java/com/kuranas/mobile/
│   │   ├── app/                    # Application, Activity raiz, ServiceLocator
│   │   ├── data/                   # implementações de repository + mappers
│   │   ├── domain/                 # entidades, portas e contratos
│   │   ├── feature/                # ownership incremental por domínio (files/images/search/settings)
│   │   ├── i18n/                   # TranslationManager
│   │   ├── infra/                  # HTTP, cache, discovery, logging, preferences
│   │   └── presentation/           # Fragments/Activities legados e comuns
│   ├── src/main/res/               # layouts XML, drawables e resources
│   └── src/main/assets/translations/
├── gradle/
├── build.gradle
└── settings.gradle
```

## Pré-requisitos

- JDK 17
- Android SDK com `compileSdk 35` instalado
- `ANDROID_HOME`/`ANDROID_SDK_ROOT` configurado

## Comandos

Build debug:

```bash
cd mobile && ./gradlew assembleDebug
```

Build release:

```bash
cd mobile && ./gradlew assembleRelease
```

Testes unitários:

```bash
cd mobile && ./gradlew test
```

Testes instrumentados:

```bash
cd mobile && ./gradlew connectedAndroidTest
```

## i18n

- Não hardcode texto visível ao usuário em Java/XML novo ou alterado.
- Traduções locais de fallback ficam em `app/src/main/assets/translations`.
- Traduções remotas são carregadas via `/api/v1/configuration/translation` por `TranslationManager`.

## Padrões de Implementação

Antes de alterar o stack mobile, seguir:

- `/docs/standards/mobile-standards.md`

Pontos obrigatórios:

- preservar API 16;
- manter Java/XML/AppCompat;
- manter separação de responsabilidades entre apresentação, domínio e dados, incluindo `feature/<domain>/{presentation,domain,data}` onde já adotado;
- executar `./gradlew test` e `./gradlew assembleDebug` nas alterações mobile.
