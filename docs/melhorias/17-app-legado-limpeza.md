# 17 — Enxugar o app legado (vira discovery + tela única)

**Tipo:** dívida técnica (demanda e-mail + kiosk) · **Prioridade:** P2 · **Depende de:** — (paralelizável com 14–16)

## Contexto

Decisão do dono (2026-06-12): o app legado (`mobile/`, Java, minSdk 16, Galaxy Tab 2 7.0) deixa de ser um cliente de navegação de mídia e vira um **painel de parede** — relógio + notificações + e-mails (task 18). O app `android/` novo já cobre navegação de arquivos/mídia. Todas as telas de Files/Imagens/Música/Vídeo/Busca/Settings saem **de vez** (não só da navegação): menos código, APK menor, app mais leve no device de 2012.

Esta task é só a remoção — preparar o terreno limpo para a task 18 construir o kiosk em cima. É paralelizável com 14–16 porque não depende de nada do backend.

## Objetivo

O app compila e roda contendo apenas: discovery/conexão de servidor (`ConnectionActivity`) → `MainActivity` hospedando uma única tela (o `HomeFragment` atual, que a task 18 transforma em kiosk). Nenhuma classe, layout, drawable ou string órfã sobra.

## O que fazer

1. Remover fragments/adapters/controllers das telas mortas e seus layouts.
2. Remover as APIs/repositórios/mappers/modelos que só elas usavam.
3. Remover o nav drawer e o coordenador de navegação.
4. Limpar `ServiceLocator`, manifest, recursos e chaves i18n órfãs.

## Como fazer

- **Remover** (código em `mobile/app/src/main/java/com/kuranas/mobile/`):
  - `presentation/{files,images,music,search,video}/` inteiros (fragments, adapters, controllers, `VideoPlayerActivity`), `presentation/home/HomeSectionAdapter.java` (listas de recentes/favoritos saem do Home).
  - `feature/{files,images,search,settings}/` inteiros.
  - `data/remote/api/{FileApi,MusicApi,SearchApi,VideoApi}.java` + repositórios, mappers e modelos de domínio correspondentes (`FileItem`, `Track`, `VideoItem`, estados de player etc. — conferir uso antes de apagar cada um).
  - `app/MainNavigationCoordinator.java`; em `activity_main.xml`, o `DrawerLayout`/lista de navegação (fica só o `content_frame`).
  - Layouts/`item_*.xml`/`view_state_*` e drawables (`ic_folder`, `ic_image`, `ic_music`, `ic_video`, `ic_search`, `ic_settings`…) que ficarem órfãos.
  - `VideoPlayerActivity` do `AndroidManifest.xml`.
- **Manter**: `ConnectionActivity` + `infra/discovery/*` (NSD/UDP/scan/validator/cache), `infra/kiosk/KioskManager`, `i18n/TranslationManager`, `infra/http/HttpClient`, `app/ServiceLocator` (enxugado), `infra/preferences/ServerPreferences`, `infra/logging/AppLogger`, `infra/cache/BitmapCache` **só se** algo restante usar (provavelmente sai também), `ConfigApi`/repositório de configuração (busca da tradução remota), `presentation/base/BaseFragment`, `presentation/home/HomeFragment` (vira a base do kiosk na 18 — pode ficar momentaneamente sem as seções de recentes/favoritos, só relógio + data).
- **`MainActivity`**: remove drawer e swap de fragments; infla direto o `HomeFragment`.
- **i18n**: remover de `assets/translations/pt-BR.json` as chaves que ficarem sem uso (NAV_*, telas mortas); manter as de discovery/erros/home. As chaves do catálogo remoto do backend não mudam (outros clients usam).
- **Dica de execução**: a task 18 vai copiar o padrão `VideoApi`/`VideoMapper`/`VideoItem` para criar `NotificationApi`/`EmailApi` — fazer a 18 ler esses arquivos no git history, ou simplesmente executar 17 e 18 em sequência por quem tiver o padrão fresco.
- **Validação**: `./gradlew assembleDebug` + testes unitários restantes; busca por referências órfãs (lint do Android Studio / `grep` por classes removidas); comparar tamanho do APK antes/depois.

## Critérios de aceite

- [ ] App compila, instala e roda: discovery → tela única com relógio. *(Compila ✓ — `assembleDebug` verde; instalar e rodar no tablet é validação manual do dono.)*
- [x] Nenhuma referência a classes/layouts/strings removidos (lint limpo, grep sem hits).
- [x] `ServiceLocator` e manifest só contêm o que restou.
- [x] Testes unitários restantes verdes.
- [x] APK menor que o atual (registrar números no commit). *(debug: 3.502.028 → 3.335.846 bytes, −4,7%.)*

**Desvio registrado na execução**: `ConfigApi`/`ConfigRepository` foram **removidos** (o "manter" do plano assumia que a tradução remota passava por eles, mas o fetch de traduções vive no próprio `TranslationManager` via `HttpClient`; o `ConfigApi` só servia a `SettingsFragment`, que morreu). Java: 92 → 23 arquivos.

## Fora de escopo

- A tela kiosk em si (task 18).
- Qualquer mudança no backend ou nos outros clients.
- Mexer no app `android/` novo.
