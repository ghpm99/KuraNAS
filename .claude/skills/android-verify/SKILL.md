---
name: android-verify
description: "Compila e roda os testes unitários de um módulo Android Gradle em background, monitorando a saída até BUILD SUCCESSFUL/FAILED, em ambiente sem IDE (JAVA_HOME explícito, local.properties com sdk.dir). Use quando o usuário disser 'compila o app', 'roda os testes do mobile', 'verifica se compila', ou após editar código Kotlin/Compose e antes de declarar sucesso."
---

# /android-verify

Verifica que um módulo Android compila e passa nos testes, num ambiente headless
(sem Android Studio, possivelmente baixa RAM). Use isto como rede de segurança
depois de editar Kotlin/Compose — não declare sucesso sem ele.

## Quando invocar

- "Compila o app / verifica se compila"
- "Roda os testes do mobile"
- Após um lote de edições Kotlin/Compose, antes de afirmar que está pronto

## Setup do ambiente (uma vez)

1. **JAVA_HOME** precisa apontar para um JDK. Neste host:
   `JAVA_HOME=/home/server/.local/jdks/jdk-17.0.19+10`. Se não souber, ache com
   `ls ~/.local/jdks` ou `which javac`.
2. **`local.properties`** com `sdk.dir` (gitignored). Se faltar:
   ```bash
   [ -f local.properties ] || echo "sdk.dir=$ANDROID_HOME" > local.properties
   ```
   (`$ANDROID_HOME` costuma ser `~/Android/Sdk`.)

## Execução (background + monitor)

Builds Gradle são lentos — **rode em background** e monitore, não bloqueie:

```bash
JAVA_HOME=<jdk> ./gradlew :app:compileDebugKotlin --console=plain 2>&1 | tail -40   # run_in_background: true
```

Depois monitore o arquivo de saída até um estado terminal. O filtro deve cobrir
**sucesso E falha** (silêncio não é sucesso):

```bash
until grep -qE "BUILD SUCCESSFUL|BUILD FAILED|FAILURE:|^e: " "$OUT"; do sleep 2; done
grep -E "BUILD SUCCESSFUL|BUILD FAILED|FAILURE:|^e: " "$OUT" | head -40
```

Para testes: `./gradlew :app:testDebugUnitTest --console=plain`.

## Disciplina

- **Compile a cada lote**, não tudo no fim — erros (import em uso removido, chave
  desbalanceada) aparecem cedo e baratos.
- `^e: ` no filtro pega os erros do compilador Kotlin linha a linha.
- **Caminhos absolutos no Bash**: o cwd pode resetar entre chamadas; prefixe com
  `cd <projeto>;` ou use caminhos absolutos.
- Mudança **visual** (layout/cor) **não** é validada por compile+testes — diga ao
  usuário que ainda convém rodar o app de verdade (veja a skill [run]/[verify]).
- Sem rede: o primeiro build pode falhar baixando dependências. Se `fetch`/download
  falhar, reporte e não insista.
