---
description: Sobe o backend Go, conecta o device por Wi-Fi (adb), recompila/instala o app e inicia o logcat
argument-hint: "[ip:porta-conexão] [ip:porta-pareamento] [código]"
allowed-tools: Bash, Read
---

Você vai preparar o ambiente de depuração do app Android do KuraNAS por Wi-Fi.
Siga o procedimento abaixo na ordem. Argumentos recebidos: `$ARGUMENTS`

## Constantes do ambiente
- Raiz do projeto: `/home/server/Documentos/Projetos/KuraNAS`
- Backend Go: pasta `backend/` — sobe com `make run` (build tag `dev`, escuta em `:8000` em todas as interfaces)
- App Android: pasta `android/` — `applicationId` debug = `com.kuranas.android.debug`, Activity = `com.kuranas.android.MainActivity`
- JDK pro Gradle: `JAVA_HOME=/home/server/.local/jdks/jdk-17.0.19+10`
- IP da máquina na LAN: descubra com `hostname -I` (o app conecta nesse IP:8000)

## Passo 0 — Coletar dados de conexão
Do `$ARGUMENTS`, interprete:
- 1º token = **IP:porta de conexão** (tela principal de "Depuração sem fio")
- 2º token (opcional) = **IP:porta de pareamento** (tela "Parear com código")
- 3º token (opcional) = **código de pareamento** de 6 dígitos

Se faltar a porta de conexão, **pergunte ao usuário** (e oriente: no device, Ajustes → Opções do desenvolvedor → Depuração sem fio; o pareamento só é necessário na 1ª vez nesta máquina). NUNCA invente IP/porta/código.

## Passo 1 — Subir o backend
- `cd /home/server/Documentos/Projetos/KuraNAS/backend`
- Se a porta 8000 já estiver ocupada, assuma que o backend já está no ar e pule o start (não derrube).
- Senão, rode `make run` em **background**, redirecionando pra `/tmp/kuranas-backend.log`.
- Aguarde a porta 8000 subir (polling de ~3s, até ~60s — `go run` compila primeiro) e valide o health:
  `curl -s http://localhost:8000/api/v1/health` deve retornar `{"status":"ok"}`.
- Mostre também o IP da LAN (`hostname -I`) pro usuário saber o que digitar no app (`<IP>:8000`).

## Passo 2 — Conectar o device via adb Wi-Fi
- Se veio pareamento (2º + 3º tokens): `echo "<código>" | adb pair <ip:porta-pareamento>`.
- Conecte: `adb connect <ip:porta-conexão>`.
- Confirme com `adb devices -l` (deve aparecer `device`, não `offline`/`unauthorized`). Se falhar, reporte e peça os dados novamente.
- Guarde o serial conectado pra usar com `adb -s <serial>` nos próximos passos.

## Passo 3 — Recompilar e instalar o app (SEMPRE atualizar)
- `cd /home/server/Documentos/Projetos/KuraNAS/android`
- Rode em **background**, redirecionando pra `/tmp/kuranas-install.log`:
  `JAVA_HOME=/home/server/.local/jdks/jdk-17.0.19+10 ./gradlew :app:installDebug`
- Aguarde até aparecer `BUILD SUCCESSFUL` ou `BUILD FAILED` no log (polling). Se falhar, mostre as linhas de erro (`^e: `) e pare.

## Passo 4 — Subir o app e iniciar o logcat
- Force-stop + relança:
  `adb -s <serial> shell am force-stop com.kuranas.android.debug`
  `adb -s <serial> logcat -c`
  `adb -s <serial> shell monkey -p com.kuranas.android.debug -c android.intent.category.LAUNCHER 1`
- Pegue o PID: `adb -s <serial> shell pidof com.kuranas.android.debug`
- Inicie o logcat filtrado por PID em **background**, gravando em `/tmp/kuranas-logcat.log`:
  `adb -s <serial> logcat --pid=<PID> -v time`

## Encerramento
Mostre um resumo em tabela: backend (porta/health), IP LAN pro app, device conectado (modelo/serial), build (sucesso), e os caminhos dos logs (`/tmp/kuranas-backend.log`, `/tmp/kuranas-install.log`, `/tmp/kuranas-logcat.log`). Lembre o usuário de digitar `<IP-LAN>:8000` na tela de conexão do app (ou usar o discovery mDNS). Deixe tudo pronto pra depurar.

> Para parar tudo depois: `fuser -k 8000/tcp` (backend), `pkill -f 'logcat --pid'` (logcat) e `adb disconnect <ip:porta-conexão>`.
