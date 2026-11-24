# 47 — Verificação local dos CAs das tasks com código pronto

**Tipo:** verificação · **Prioridade:** P0 · **Spec:** as de cada task listada

## Contexto

Até 2026-11-17 as tasks com código pronto e `make ci` verde ficavam `🚫 bloqueada` esperando o dono validar os `CA` no ambiente dele, o que travava as dependentes. Decisão do dono em 2026-11-17: não bloquear; o agente verifica no sistema rodando **localmente** e o dono reporta defeitos que encontrar no uso. As tasks abaixo passaram a `✅ concluída (código)` e os `CA` desmarcados delas são verificados aqui.

## O que fazer

Para cada task, abrir o arquivo dela, pegar os `CA` desmarcados e verificar com o backend local (`cd backend && make run`, Postgres local de dev), chamadas à API (curl) e consultas ao banco local; o que for de tela, pelo frontend local (`cd frontend && yarn dev`) quando a API não bastar. Marcar `[x]` no arquivo da task no mesmo commit da verificação, com a evidência resumida (comando e resultado). Defeito achado vira correção na própria task (commit `fix`) antes de marcar. Sem aparelho: Android/tablet ficam com build + testes unitários e o que der para verificar pela API — anotar no CA o que não pôde ser exercitado.

- [x] 20 — paginação web (`has_next`) — pasta `teste-paginacao-task20`.
- [x] 21 — falha ao carregar mais não esvazia a lista.
- [x] 22 — Android: descoberta do servidor (API: UDP/mDNS respondem localmente).
- [x] 23 — Android: explorer paginado (API das páginas).
- [ ] 25 — excluir captura não apaga o vídeo.
- [x] 26 — Takeout: estrutura e sem sobrescrita (zip de Takeout montado para teste).
- [x] 29 — preferências de indexação com efeito (`21/CA-12`).
- [x] 30 — reindexar por escopo.
- [x] 32 — atualização confere checksum (pacote adulterado recusado).
- [x] 33 — desligamento: aviso e cancelar.
- [x] 35 — envio em partes web (arquivo grande, queda de conexão simulada, cancelar, já existia, conflito).
- [x] 40 — correção de metadados de música (verificada na própria task).
- [x] 06, 11, 17, 18, 19 — o que der para verificar localmente (watcher fsnotify no Linux, WebDAV montado com davfs2/curl, app legado e captura só por build/testes).

## Critérios de aceite

- [ ] Cada item acima com os `CA` da task marcados ou anotados com o motivo de não poder ser exercitado localmente.
- [ ] `make ci` verde ao final.
