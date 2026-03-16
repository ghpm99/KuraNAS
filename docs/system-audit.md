# Auditoria Técnica Completa do Sistema KuraNAS

## 1. Objetivo e escopo

Este documento descreve o estado real do sistema com base no código-fonte presente neste repositório em `2026-03-16`.

O foco desta auditoria é cobrir:

- arquitetura geral;
- fluxo de startup;
- composição backend/frontend;
- persistência e modelo de dados;
- workers, jobs e indexação;
- funcionalidades por domínio;
- configuração e comportamento em runtime;
- atualização automática;
- métodos reais de backup e restauração possíveis;
- limitações, lacunas e discrepâncias observáveis no código;
- referência cruzada com as rotas expostas.

Este documento deve ser lido junto com [system-api-reference.md](/mnt/wsl/PHYSICALDRIVE4/projects/KuraNAS/docs/system-api-reference.md).

## 2. Resumo executivo

KuraNAS é um sistema NAS pessoal com backend em Go e frontend em React.

O backend:

- expõe uma API HTTP em `/api/v1`;
- mantém catálogo de arquivos, metadados e estados de reprodução;
- executa workers assíncronos para varredura, checksum, thumbnails, jobs e playlists de vídeo;
- carrega traduções JSON compartilhadas com o frontend;
- serve também o frontend compilado.

O frontend:

- é uma SPA React + Vite;
- consome a API via `axios` usando `getApiV1BaseUrl()`;
- organiza a experiência em domínios: Home, Files, Favorites, Images, Music, Videos, Analytics, Settings, About e Activity Diary;
- usa React Query para estado de servidor e Context para estado local/domínio.

## 3. Estrutura real do sistema

### 3.1 Monorepo

- `backend/`: API, workers, banco, i18n e scripts de extração de metadados.
- `frontend/`: SPA React.
- `docs/`: padrões e documentação.
- `build/`: artefatos do empacotamento integrado.

### 3.2 Backend

Principais áreas:

- `backend/cmd/nas`: entrypoints.
- `backend/internal/app`: bootstrap, contexto de aplicação e rotas.
- `backend/internal/config`: carga de ambiente e parâmetros de runtime.
- `backend/internal/api/v1`: domínios HTTP.
- `backend/internal/worker`: processamento assíncrono.
- `backend/pkg/database`: conexão, contexto e migrations.
- `backend/pkg/i18n`: carregamento de traduções.
- `backend/pkg/logger`: log transacional.
- `backend/scripts`: extração Python de metadados.

### 3.3 Frontend

Principais áreas:

- `frontend/src/app`: `App.tsx` e definição de rotas.
- `frontend/src/pages`: wrappers de página.
- `frontend/src/components`: telas, layouts, hooks e providers.
- `frontend/src/service`: cliente HTTP por domínio.
- `frontend/src/types`: contratos do frontend.

## 4. Arquitetura funcional

### 4.1 Backend em camadas

O padrão dominante é:

- `Handler -> Service -> Repository`

Responsabilidades:

- `Handler`: parse de request, status HTTP, serialização.
- `Service`: regra de negócio, transação, orquestração.
- `Repository`: SQL e mapeamento.

### 4.2 Contexto central de aplicação

`backend/internal/app/context.go` monta um `AppContext` único com:

- DB;
- logger;
- fila global de tasks;
- contextos de `Files`, `Jobs`, `Diary`, `Music`, `Video`, `Analytics`, `Configuration`, `Search`;
- `UpdateService` e `UpdateHandler`.

### 4.3 Frontend em providers

`AppProviders` compõe:

- `QueryClientProvider`;
- `I18nProvider`;
- `SettingsProvider`;
- `SnackbarProvider`;
- `ThemeProvider`;
- `BrowserRouter`;
- `GlobalSearchProvider`.

O `GlobalMusicProvider` é aplicado em `App.tsx` ao redor do conteúdo da aplicação.

## 5. Fluxo detalhado de startup

## 5.1 Startup backend em desenvolvimento

O entrypoint `backend/cmd/nas/main.go`:

1. escreve log de início;
2. chama `app.InitializeApp()`;
3. cria `signal.NotifyContext` para `SIGINT` e `SIGTERM`;
4. sobe o servidor HTTP em `:8000`;
5. em caso de sinal, chama `application.Stop()`;
6. aguarda encerramento gracioso por até 6 segundos.

## 5.2 Startup backend em Windows não-dev

O entrypoint `backend/cmd/nas/main_windows.go`:

1. registra o programa como serviço Windows via `kardianos/service`;
2. ajusta o working directory para a pasta do executável;
3. cria log em arquivo dentro de `log/`;
4. inicializa a aplicação;
5. injeta um `shutdownFn` no updater para encerrar o serviço após update;
6. sobe HTTP em `:8000`;
7. suporta comandos de serviço: `install`, `uninstall`, `start`, `stop`.

Observação importante:

- não existe entrypoint Linux não-dev no repositório atual;
- há configuração Linux de paths e suporte parcial no updater, mas o caminho de execução/empacotamento principal do repositório está orientado a Windows.

## 5.3 Etapas internas de `InitializeApp()`

Ordem real:

1. `LoadConfig()`
2. `InitializeConfig()`
3. `LoadTranslations()`
4. ajuste do `gin` mode conforme `ENV`
5. `ConfigDatabase()`
6. criação do `AppContext`
7. criação do router `gin.Default()`
8. registro de rotas
9. montagem do `WorkerContext`
10. `StartWorkers()`

### 5.3.1 Carga de configuração

`LoadConfig()` tenta carregar `.env` no path resolvido por `GetBuildConfig("EnvFilePath")`.

Se não existir, o sistema segue usando variáveis do ambiente do processo.

### 5.3.2 Inicialização de `AppConfig`

Campos efetivos:

- `ENTRY_POINT`
- `LANGUAGE`
- `ENABLE_WORKERS`
- `ENV`
- `DB_HOST`
- `DB_PORT`
- `DB_USER`
- `DB_PASSWORD`
- `DB_NAME`
- `ALLOWED_ORIGINS`
- `WORKER_CONCURRENCY_CHECKSUM`
- `WORKER_CONCURRENCY_METADATA`
- `WORKER_CONCURRENCY_THUMBNAIL`
- `WORKER_RETRY_BACKOFF_MS`
- `WORKER_SCHEDULER_POLL_MS`
- `WORKER_MAX_CONCURRENT_JOBS`

Campos derivados:

- `StartupTime`
- `RecentFilesKeep = 10`

### 5.3.3 Carregamento de traduções

O backend carrega o arquivo JSON do locale atual a partir de:

- path de build/instalação quando existir;
- fallback para `translations/` ou `backend/translations/`.

O frontend consome o mesmo conteúdo via `/configuration/translation`.

### 5.3.4 Banco de dados

`ConfigDatabase()` hoje abre conexão com driver `postgres`.

Fluxo:

1. monta DSN com `host`, `port`, `user`, `dbname`, `password`, `sslmode=disable`;
2. abre conexão;
3. executa `migrations.Init(db)`.

Discrepância importante:

- o repositório mantém referências antigas a SQLite (`DbPath`, dependência `go-sqlite3`, testes em memória);
- porém o bootstrap de produção/desenvolvimento mostrado pelo código usa PostgreSQL, não SQLite.

### 5.3.5 Migrations na inicialização

As migrations são transacionais e idempotentes:

- cria tabela `migrations`;
- registra migrations já aplicadas;
- executa apenas as pendentes;
- faz `commit` ao final.

### 5.3.6 Registro de rotas

As rotas são agrupadas por domínio:

- `files`
- `diary`
- `music`
- `video`
- `analytics`
- `jobs`
- `configuration`
- `update`
- `search`

Além disso:

- `/assets` serve o frontend compilado;
- `NoRoute` responde `./dist/index.html`.

### 5.3.7 Inicialização dos workers

Se `ENABLE_WORKERS != true`, os workers não sobem.

Se habilitados:

1. monta `JobScheduler`;
2. monta `JobOrchestrator`;
3. inicia scheduler;
4. cria `numWorkers` goroutines consumidores da fila legacy;
5. agenda varredura inicial;
6. inicia watcher periódico do `ENTRY_POINT`.

## 5.4 Startup frontend

Fluxo real:

1. `frontend/src/main.tsx` importa `index.css`;
2. cria root React;
3. renderiza `App`;
4. `AppProviders` sobe Query Client, i18n, settings, snackbar, tema e router;
5. `App.tsx` registra rotas da SPA;
6. `GlobalMusicProvider` sobe o player global.

### 5.4.1 Resolução da URL da API

Ordem real:

1. `globalThis.__KURANAS_API_URL__`
2. `VITE_API_URL`
3. `process.env.VITE_API_URL`
4. fallback para `/api/v1`

## 6. Modelo de dados persistido

## 6.1 Tabelas centrais

### `home_file`

Catálogo principal do sistema.

Campos principais:

- `id`
- `name`
- `path`
- `parent_path`
- `format`
- `size`
- `updated_at`
- `created_at`
- `last_interaction`
- `last_backup`
- `type`
- `checksum`
- `deleted_at`
- `starred` (adicionado por migration posterior)

Índices:

- por `path`
- por `parent_path`
- por `name`
- por `(path, name)`

### `image_metadata`

Metadados ricos de imagem:

- dimensões;
- DPI;
- EXIF;
- GPS;
- câmera/lente;
- descrição/autoria;
- classificação `classification_category` e `classification_confidence`.

### `audio_metadata`

Metadados de áudio:

- mime;
- duração;
- bitrate;
- sample rate;
- channels;
- title;
- artist;
- album;
- album artist;
- track number;
- genre;
- year;
- letras e editoriais.

### `video_metadata`

Metadados de vídeo:

- formato/container;
- tamanho;
- duração;
- largura/altura;
- frame rate;
- codec;
- aspecto;
- dados de áudio embutido.

### `recent_file`

Relaciona IP do cliente com arquivos acessados recentemente.

### `playlist`

Playlists musicais persistidas.

### `playlist_track`

Relação ordenada entre playlist musical e arquivos.

### `player_state`

Estado do player de música por `client_id`.

### `video_playlist`

Playlists/contextos de vídeo.

Campos relevantes:

- `type`
- `source_path`
- `name`
- `is_hidden`
- `is_auto`
- `group_mode`
- `classification`
- `last_played_at`

### `video_playlist_item`

Itens ordenados da playlist de vídeo.

### `video_playback_state`

Estado do player de vídeo por `client_id`.

### `video_playlist_exclusion`

Exclusões manuais para playlists automáticas.

### `video_behavior_event`

Histórico comportamental:

- `started`
- `paused`
- `resumed`
- `completed`
- `skipped`
- `abandoned`

### `worker_job`

Job assíncrono.

### `worker_step`

Passos de um job com dependências e payload JSON.

### `activity_diary`

Diário de atividade.

### `app_settings`

Documento JSON de preferências persistidas.

## 6.2 Observação importante sobre backup

O campo `last_backup` existe em `home_file`, mas o código atual não implementa um fluxo de backup/restauração que atualize ou opere esse campo.

Na prática:

- o sistema indexa arquivos;
- o sistema não executa backup desses arquivos.

## 7. Configuração

## 7.1 Configuração por ambiente

### Obrigatórias na prática

- `ENTRY_POINT`: raiz observada/indexada.
- `DB_HOST`
- `DB_PORT`
- `DB_USER`
- `DB_PASSWORD`
- `DB_NAME`

### Comportamentais

- `LANGUAGE`
- `ENABLE_WORKERS`
- `ALLOWED_ORIGINS`
- `ENV`

### De tuning de worker

- `WORKER_CONCURRENCY_CHECKSUM`
- `WORKER_CONCURRENCY_METADATA`
- `WORKER_CONCURRENCY_THUMBNAIL`
- `WORKER_RETRY_BACKOFF_MS`
- `WORKER_SCHEDULER_POLL_MS`
- `WORKER_MAX_CONCURRENT_JOBS`

## 7.2 Preferências persistidas em `/configuration/settings`

Categorias:

- `library`
- `indexing`
- `players`
- `appearance`
- `language`

### `library`

- `runtime_root_path`
- `watched_paths`
- `remember_last_location`
- `prioritize_favorites`

### `indexing`

- `workers_enabled` (somente leitura, derivado de `ENABLE_WORKERS`)
- `scan_on_startup`
- `extract_metadata`
- `generate_previews`

### `players`

- `remember_music_queue`
- `remember_video_progress`
- `autoplay_next_video`
- `image_slideshow_seconds`

### `appearance`

- `accent_color`: `violet`, `cyan`, `rose`
- `reduce_motion`

### `language`

- `current`
- `available`

## 7.3 O que é efetivamente aplicado hoje

Aplicado de fato:

- idioma em runtime;
- cor de destaque no frontend;
- `reduce_motion` no frontend;
- hidratação de fila musical via `remember_music_queue`;
- persistência de progresso de vídeo via `remember_video_progress`;
- autoplay do próximo vídeo via `autoplay_next_video`;
- intervalo de slideshow de imagens.

Persistido, mas não claramente aplicado ao fluxo principal de indexação:

- `watched_paths`;
- `remember_last_location`;
- `prioritize_favorites`;
- `scan_on_startup`;
- `extract_metadata`;
- `generate_previews`.

Motivo:

- o pipeline principal continua tomando decisões a partir de `ENTRY_POINT` e `ENABLE_WORKERS`;
- não há uso operacional desses flags no backend além da leitura/persistência de settings.

## 8. Fluxos operacionais detalhados

## 8.1 Fluxo de indexação no startup

Quando workers estão habilitados:

1. `StartWorkers()` sobe scheduler e orquestrador;
2. `startWorkersScheduler()` chama `enqueueStartupScanJob()`;
3. o job `startup_scan` recebe 3 steps:
   - `scan_filesystem`
   - `diff_against_db`
   - `mark_deleted`
4. o scheduler processa jobs conforme prioridade e disponibilidade;
5. o catálogo é reconciliado com o filesystem.

## 8.2 Fluxo de watcher do filesystem

O watcher atual:

- faz snapshot periódico do `ENTRY_POINT` a cada 5 segundos;
- compara mtime/tamanho/tipo;
- usa debounce de 2 segundos.

Se houver muitas mudanças:

- agenda job de varredura completa.

Se houver poucas mudanças:

- exclusão vira job `mark_deleted`;
- arquivo novo/modificado vira job de processamento do arquivo.

## 8.3 Fluxo de upload

Fluxo real:

1. frontend envia multipart para `/files/upload`;
2. backend valida `target_path` dentro do `ENTRY_POINT`;
3. grava os arquivos fisicamente;
4. cria job `upload_process`;
5. cria steps por arquivo:
   - `persist`
   - `metadata`
   - `checksum`
   - `thumbnail` para imagem/vídeo
   - `playlist_index` para vídeo
6. resposta HTTP retorna `202 Accepted` com `job_id`.

## 8.4 Fluxo de persistência do catálogo

Na camada `files`:

- cria ou atualiza `home_file`;
- faz upsert do metadata especializado;
- pode atualizar `starred`, `checksum`, `deleted_at`.

## 8.5 Fluxo de thumbnails e previews

Arquivos de imagem:

- thumbnail PNG.

Arquivos de vídeo:

- thumbnail PNG;
- preview GIF animado.

Se o arquivo não existir mais em disco:

- handlers retornam `404`.

## 8.6 Fluxo de recentes

Ao servir blob por `/files/blob/:id`:

1. backend resolve o arquivo;
2. registra acesso por IP;
3. poda o histórico antigo conforme `RecentFilesKeep`.

## 8.7 Fluxo de música

### Biblioteca

O sistema monta catálogos derivados com base em `audio_metadata`:

- artistas;
- álbuns;
- gêneros;
- pastas;
- todas as faixas.

### Playlists musicais

Tipos reais:

- manuais;
- de sistema;
- automáticas.

Playlists automáticas negativas:

- `-1`: Continue Listening
- `-2`: Recently Added
- `-3`: Favorites

Essas playlists são somente leitura.

### Queue global

O `GlobalMusicProvider`:

- mantém a fila local;
- toca via stream `/files/stream/:id`;
- sincroniza estado com backend;
- pode reidratar fila anterior a partir de `player_state` + `now-playing`.

## 8.8 Fluxo de vídeo

### Playback contextual

Ao iniciar um vídeo:

- se vier `playlist_id`, o sistema valida pertencimento;
- se não vier, cria ou reutiliza playlist contextual por pasta.

### Estado de reprodução

Persistido por `client_id`:

- playlist atual;
- vídeo atual;
- tempo atual;
- duração;
- pausado;
- concluído.

### Navegação

Existem comandos:

- próximo;
- anterior;
- update de estado;
- tracking de evento comportamental.

### Playlists inteligentes

`RebuildSmartPlaylists()`:

1. lê todos os vídeos com metadados;
2. lê eventos comportamentais;
3. executa `PlaylistEngine`;
4. produz grupos automáticos;
5. upserta playlists automáticas;
6. remove itens automáticos antigos;
7. reinsere itens respeitando exclusões manuais.

### Catálogo Home de vídeo

Seções reais:

- `continue`
- `series`
- `movies`
- `personal`
- `recent`

## 8.9 Fluxo de imagens

O provider de imagens:

- busca `/files/images`;
- pagina em lotes;
- suporta agrupamento por `date`, `type`, `name`.

Na UI:

- há seções derivadas de navegação;
- há coleções automáticas por álbum;
- há coleções por pasta;
- o viewer suporta zoom, filmstrip, slideshow e favoritar;
- uma imagem pode ser aberta diretamente por `?image=<id>&imagePath=<path>`.

## 8.10 Fluxo de analytics

O backend consolida:

- storage total/usado/livre/crescimento;
- total de arquivos/pastas;
- série temporal;
- distribuição por tipo e extensão;
- hot folders;
- top folders;
- arquivos recentes;
- grupos duplicados;
- cobertura de metadata;
- fila/processamento pendente;
- saúde do indexador.

Períodos suportados:

- `24h`
- `7d`
- `30d`
- `90d`

## 8.11 Fluxo de global search

Atalho de UI:

- `Ctrl+K` no Windows/Linux;
- `Cmd+K` no macOS.

A busca:

1. faz debounce natural pelo `useDeferredValue`;
2. consulta `/search/global` com mínimo de 2 caracteres;
3. combina resultados de:
   - arquivos
   - pastas
   - artistas
   - álbuns
   - playlists musicais
   - playlists de vídeo
   - vídeos
   - imagens
4. converte resultados em navegação contextual.

## 8.12 Fluxo do diário de atividade

O diário permite:

- criar atividade;
- listar atividades;
- obter resumo recente;
- duplicar atividade;
- atualizar atividade.

Comportamento específico:

- ao criar uma nova entrada, a entrada anterior aberta é encerrada.

Observação:

- o handler de update atual usa `PostForm("data")` em vez de payload JSON estruturado;
- isso indica contrato mais frágil do que os demais domínios.

## 9. Funcionalidades por área da aplicação

## 9.1 Home

Agrega:

- analytics resumido;
- favoritos;
- imagens recentes;
- retomada de vídeo;
- retomada de música;
- seções iniciais de vídeo.

## 9.2 Files

Capacidades:

- árvore de arquivos;
- navegação por URL de path;
- breadcrumb;
- grid/list view;
- favoritos;
- upload;
- criar pasta;
- mover;
- copiar;
- renomear;
- deletar;
- reindexar manualmente;
- download blob;
- thumbs;
- preview de vídeo;
- stream de áudio e vídeo.

## 9.3 Favorites

Lista arquivos com categoria `starred`.

## 9.4 Images

Capacidades:

- biblioteca de imagens;
- recentes;
- capturas;
- fotos;
- por pasta;
- álbuns automáticos;
- favoritar;
- slideshow;
- navegação direta por query string.

## 9.5 Music

Capacidades:

- home musical;
- playlists manuais;
- playlists automáticas;
- artistas;
- álbuns;
- gêneros;
- pastas;
- fila global;
- now playing;
- persistência de estado do player.

## 9.6 Videos

Capacidades:

- home catalogado;
- continue watching;
- playlists automáticas e manuais;
- ocultar playlist;
- renomear playlist;
- reordenar itens;
- adicionar/remover vídeos;
- player dedicado;
- contexto de origem;
- autoplay opcional.

## 9.7 Analytics

Capacidades:

- overview operacional;
- visão de biblioteca;
- saúde do indexador;
- duplicatas;
- capacidade;
- distribuição por tipo/extensão.

## 9.8 Settings

Capacidades:

- editar preferências persistidas;
- trocar idioma;
- trocar cor de destaque;
- reduzir motion;
- editar watched paths;
- ajustar flags de player;
- visualizar root e estado de workers.

## 9.9 About

Capacidades:

- expor versão;
- commit;
- plataforma;
- path monitorado;
- idioma atual;
- workers habilitados;
- horário de startup;
- versões de Go e Node.

## 9.10 Jobs

Capacidades:

- listar jobs;
- inspecionar um job;
- listar steps;
- cancelar job.

Hoje não há tela clara no frontend consumindo esse domínio de forma visível para o usuário final.

## 9.11 Auto-update

Capacidades no backend:

- consultar latest release no GitHub;
- comparar versão atual vs última versão;
- baixar asset por SO;
- validar tamanho;
- extrair zip;
- substituir binário e diretórios de runtime.

Hoje não há tela clara no frontend consumindo isso, embora exista `frontend/src/service/update.ts`.

## 10. Atualização automática

## 10.1 Endpoint de status

`GET /api/v1/update/status`

Retorna:

- versão atual;
- última versão;
- flag de update;
- URL da release;
- data;
- notas;
- nome do asset;
- tamanho do asset.

## 10.2 Endpoint de aplicação

`POST /api/v1/update/apply`

Fluxo:

1. consulta latest release no GitHub;
2. escolhe asset por SO:
   - `kuranas-windows.zip`
   - `kuranas-linux.zip`
3. baixa para pasta temporária;
4. confere tamanho do arquivo;
5. extrai zip com proteção contra zip-slip;
6. descobre diretório de instalação a partir do executável;
7. renomeia binário atual para `.old`;
8. copia novo binário;
9. substitui `dist`, `icons`, `translations`, `scripts`;
10. preserva `scripts/.venv` quando existir;
11. agenda shutdown do processo se houver callback registrado.

## 10.3 Observações operacionais

- o updater depende de acesso HTTP ao GitHub;
- a aplicação não faz rollback completo dos assets não-binários;
- existe proteção de restauração apenas no binário principal;
- a integração de shutdown pós-update está explícita no entrypoint Windows.

## 11. Backup e restauração

## 11.1 O que existe no código

Não existe módulo de backup/restauração dedicado.

Não existem:

- endpoints de backup;
- agendamento de backup;
- restore automatizado;
- versionamento de snapshots;
- retenção automática;
- restauração de banco via UI.

## 11.2 O que precisa ser preservado manualmente

Para backup operacional real do sistema, devem ser preservados:

- banco PostgreSQL usado pelo backend;
- árvore de arquivos monitorada em `ENTRY_POINT`;
- `.env` da instalação;
- `translations/` se customizado;
- `scripts/.venv` se o ambiente Python for provisionado localmente;
- `dist/`, `icons/`, `scripts/` quando o objetivo for backup completo da instalação, não só dos dados.

## 11.3 Método recomendado de backup lógico

Banco:

- dump do PostgreSQL da base configurada por `DB_NAME`.

Conteúdo:

- cópia recursiva do `ENTRY_POINT`.

Configuração:

- cópia do `.env`.
- cópia opcional de `app_settings` no dump do banco.

## 11.4 Método recomendado de restauração

1. restaurar a base PostgreSQL;
2. restaurar a árvore de arquivos no mesmo `ENTRY_POINT`;
3. restaurar `.env`;
4. subir a aplicação;
5. executar reindexação se necessário para reconciliar catálogo e disco.

## 11.5 Limitação importante

O sistema não garante consistência transacional entre:

- estado do banco;
- estado do filesystem monitorado.

Portanto, restauração parcial pode exigir nova varredura.

## 12. CORS, assets e serving do frontend

O backend aplica CORS com:

- origins vindos de `ALLOWED_ORIGINS`;
- métodos `GET`, `PUT`, `POST`, `DELETE`;
- credenciais habilitadas.

O frontend compilado é servido por:

- `/assets`
- fallback SPA com `NoRoute -> ./dist/index.html`

## 13. Build e empacotamento

## 13.1 Build integrado na raiz

`make`:

1. build frontend;
2. build backend;
3. move artefatos para `build/`;
4. chama `deploy` local via `Makefile.local`.

Itens copiados:

- `frontend/dist`
- binário backend
- `icons`
- `translations`

Observação:

- o `Makefile` raiz não copia `scripts/` para `build/`, embora o updater trate `scripts/` como asset atualizável.

## 13.2 Build backend

O `backend/Makefile` atual gera binário Windows:

- `GOOS=windows`
- saída `kuranas.exe`

## 14. Discrepâncias e limitações relevantes da auditoria

### 14.1 Banco

- documentação histórica menciona SQLite/PostgreSQL;
- bootstrap real usa PostgreSQL;
- SQLite permanece em testes e helpers de path, não como runtime principal.

### 14.2 Linux

- há paths Linux e nome de asset Linux no updater;
- não há entrypoint Linux não-dev no repositório atual;
- build principal do backend está orientado a Windows.

### 14.3 Settings de indexing

Persistem na configuração, mas não governam efetivamente o pipeline principal:

- `scan_on_startup`
- `extract_metadata`
- `generate_previews`

### 14.4 Settings de biblioteca

Persistem na configuração, mas não governam efetivamente a raiz observada:

- `watched_paths`
- `remember_last_location`
- `prioritize_favorites`

O runtime continua centrado em `ENTRY_POINT`.

### 14.5 Auto-update no frontend

- existe cliente HTTP para update;
- não há tela evidente no frontend consumindo essa funcionalidade.

### 14.6 Backup

- não existe subsistema de backup real;
- `last_backup` hoje é apenas dado persistido, não fluxo operacional.

## 15. Mapa de navegação frontend

Rotas principais:

- `/home`
- `/files/*`
- `/favorites`
- `/settings`
- `/internal/activity-diary`
- `/analytics/*`
- `/about`
- `/images/*`
- `/music/*`
- `/videos/*`
- `/video/:id`

Legados/redirecionamentos:

- `/starred -> /favorites`
- `/activity-diary -> /internal/activity-diary`
- `/ -> /home`

## 16. Encerramento

O sistema é funcionalmente forte em:

- catálogo de arquivos;
- mídia local;
- indexação assíncrona;
- playlists e playback;
- analytics de biblioteca;
- busca global;
- configurações persistidas.

Os pontos mais importantes para operação e evolução são:

- alinhar a documentação e o bootstrap do banco;
- decidir se `watched_paths` e flags de indexing devem realmente controlar o worker;
- decidir se auto-update será exposto na UI;
- implementar backup/restauração nativos caso isso seja requisito do produto.

Para a referência rota a rota, consulte [system-api-reference.md](/mnt/wsl/PHYSICALDRIVE4/projects/KuraNAS/docs/system-api-reference.md).
