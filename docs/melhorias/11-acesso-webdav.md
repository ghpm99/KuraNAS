# 11 — Acesso WebDAV (montar o NAS como unidade de rede)

**Tipo:** feature de maturidade · **Prioridade:** P3

## Contexto

Hoje o único acesso aos arquivos é a API HTTP própria (`/api/v1`), consumida pelos clientes do projeto. Nenhum dispositivo de fora do ecossistema consegue **montar** o armazenamento: o Explorer do Windows, o Finder, gerenciadores de arquivos Android, players de mídia (Kodi/VLC) e ferramentas de backup esperam um protocolo padrão — SMB, NFS ou WebDAV. É a diferença mais visível entre o KuraNAS e qualquer NAS estabelecido: neles, protocolos de rede são o núcleo do produto.

Dos três, **WebDAV é o único viável de embutir** no binário Go (`golang.org/x/net/webdav`, biblioteca madura, sem CGO, sem privilégios de SO). SMB/NFS exigiriam reimplementar servidores complexos ou depender de recursos do sistema operacional — fora do alcance razoável do projeto.

**Pré-requisito: task 04 (whitelist de IPs).** Expor WebDAV sem nenhum controle de acesso agravaria o problema de segurança atual — o `/dav` deve nascer atrás do mesmo middleware de whitelist.

## Objetivo

O usuário monta o KuraNAS como unidade de rede (Windows "Mapear unidade de rede", `davfs2`/GNOME/KDE no Linux, apps de arquivos no Android) com leitura e escrita, a partir de qualquer dispositivo com IP cadastrado na whitelist, e as mudanças feitas por WebDAV aparecem no índice como qualquer outra mudança no disco.

## O que fazer

1. Embutir um servidor WebDAV servindo as raízes de armazenamento sob um prefixo próprio (ex.: `/dav/`).
2. Garantir que o `/dav` fica atrás do middleware de whitelist da task 04 (sem credencial — acesso anônimo restrito por IP, conforme a decisão de não ter autenticação).
3. Garantir que escritas via WebDAV entram no índice (via watcher) e que a lixeira/regras de path são respeitadas.

## Como fazer

- **Servidor**: `golang.org/x/net/webdav` com `webdav.Handler{FileSystem: webdav.Dir(raiz), LockSystem: webdav.NewMemLS()}`. Montar no Gin via `router.Any("/dav/*path", gin.WrapH(handler))` — atenção: registrar fora do grupo com middlewares que interferem (gzip pode corromper PUT/PROPFIND; CORS é irrelevante para clientes nativos).
- **Controle de acesso**: nenhum login — o middleware global de whitelist de IPs (task 04) já cobre o `/dav`, desde que registrado no engine antes das rotas (conferir que o `gin.WrapH` do WebDAV não é montado fora do alcance do middleware). Acesso anônimo restrito por IP evita inclusive a limitação do Windows com HTTP Basic sem HTTPS — montar `http://servidor:8000/dav/` funciona sem credencial.
- **Raízes**: com a task 10 feita, expor cada raiz como diretório de topo (`/dav/<label>/...`) usando um `webdav.FileSystem` próprio que despacha pela primeira parte do path; sem a task 10, servir o `ENTRY_POINT` direto.
- **Indexação**: nenhuma integração especial — escritas WebDAV tocam o disco e o watcher (task 03/06) indexa. Validar esse caminho num teste manual: criar arquivo via cliente WebDAV → aparece na aba de arquivos.
- **Exclusões**: esconder `.kuranas-trash/` (task 09) e quaisquer diretórios internos do filesystem WebDAV exposto.
- **Config**: liga/desliga via configuração (`WEBDAV_ENABLED` ou flag na tabela de configuração), default desligado.
- **Testes**: handler com `litmus`-style básico via testes Go (PROPFIND, GET, PUT, MKCOL, DELETE, MOVE) usando cliente HTTP de teste; whitelist (IP fora da lista → 403 em qualquer verbo de `/dav/`). Validação manual com Explorer (Windows) e `davfs2`/Nautilus (Linux), documentada na task.

## Critérios de aceite

- [ ] Com WebDAV habilitado e a partir de IP cadastrado, é possível montar a unidade no Windows Explorer e num cliente Linux, navegar, abrir, criar, renomear e excluir arquivos.
- [ ] De IP fora da whitelist, qualquer verbo em `/dav/` responde 403.
- [ ] Arquivo criado via WebDAV aparece na aba de arquivos (e nas abas de mídia, quando aplicável) sem ação manual.
- [ ] Diretórios internos (lixeira) não aparecem nem são acessíveis via WebDAV.
- [ ] Com WebDAV desabilitado (default), `/dav/` não existe.
- [ ] Download/upload de arquivo grande (>1 GB) funciona sem carregar o arquivo inteiro em memória.
- [ ] `make ci-backend` verde (cobertura ≥ 80%).

## Fora de escopo

- SMB, NFS, FTP/SFTP, DLNA — decisão registrada: não embutir; quem precisar de SMB usa o compartilhamento do próprio Windows sobre as mesmas pastas.
- Permissões por pasta/usuário e qualquer forma de credencial (decisão registrada na task 04: sem autenticação; controle é por IP).
- Locking distribuído persistente (o `MemLS` em memória é suficiente; locks não sobrevivem a restart).
- Otimizações de cache/ETag além do que a biblioteca oferece.
