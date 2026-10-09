---
name: create-skill
description: "Cria uma nova skill do Claude Code seguindo o padrão oficial (Agent Skills): diretório com SKILL.md, frontmatter name+description bem escrito para disparo automático, e corpo de instruções/checklist. Use quando o usuário disser 'cria uma skill para X', 'transforma esse procedimento em skill', 'nova skill', ou invocar /create-skill. Baseada em code.claude.com/docs/en/skills."
---

# /create-skill

Cria uma skill nova, bem-formada, a partir de um procedimento repetível. Uma boa
skill captura **passos + armadilhas** — escreva quando você se pegar repetindo a
mesma checklist no chat, ou quando um trecho do CLAUDE.md virou procedimento (não fato).

## Quando invocar

- "Cria uma skill para X" / "Transforma esse procedimento em skill"
- "Nova skill" / `/create-skill`
- Após perceber um fluxo repetitivo nesta sessão que vale reusar

## Onde a skill vive (define o comando)

O comando vem do **nome do diretório**, não do frontmatter:

| Local | Comando | Escopo |
|---|---|---|
| `~/.claude/skills/<dir>/SKILL.md` | `/<dir>` | pessoal (todos os projetos) |
| `<repo>/.claude/skills/<dir>/SKILL.md` | `/<dir>` | projeto (versionado) |
| plugin `skills/<dir>/SKILL.md` | `/<plugin>:<dir>` | plugin |

Use **pessoal** para algo que serve em qualquer projeto seu; **projeto** quando é
específico do repo e deve ser versionado com ele. Nomeie o diretório em
`kebab-case`, curto e específico (vira o `/comando`).

## Frontmatter (entre `---`)

| Campo | Obrigatório | Para quê |
|---|---|---|
| `name` | não | rótulo de exibição (default: nome do diretório) |
| `description` | **recomendado** | o que faz **e quando usar**; Claude usa para decidir disparar |
| `when_to_use` | não | gatilhos/exemplos extras; somado à description |
| `disable-model-invocation` | não | `true` = só o usuário invoca (`/nome`); use para fluxos com efeito colateral |
| `allowed-tools` | não | ferramentas liberadas sem pedir permissão enquanto ativa |
| `arguments` | não | argumentos nomeados (`$nome` no corpo) |

Limite: `description` + `when_to_use` são truncados em **~1.536 caracteres** na
listagem — seja conciso e ponha o caso-chave primeiro.

## Como escrever a `description` (o que mais importa)

É o que faz a skill disparar na hora certa. Regras:
- **Caso de uso primeiro**, depois os gatilhos.
- Terceira pessoa, descrevendo o que a skill faz: "Migra telas...", "Detecta e remove...".
- **Inclua frases-gatilho reais** que o usuário diria: `Use quando o usuário disser
  'X', 'Y', 'Z'`. (Estilo já usado nas skills deste workspace.)
- Se útil, diga também **quando NÃO** usar (evita disparo errado).

Exemplo bom:
> "Detecta e remove imports órfãos após refactor Kotlin... Use quando o usuário
> disser 'limpa os imports não usados', 'ficou código morto?', ou após um refactor grande."

## Corpo (markdown após o frontmatter)

Mantenha **< 500 linhas**. Estruture com:
1. Uma frase do que a skill faz e quando NÃO usar.
2. `## Quando invocar` — bullets de gatilho.
3. `## Passos` / `## Procedimento` — numerados, com comandos concretos.
4. `## Armadilhas` — **a parte de maior valor**: os gotchas que você só descobre
   errando (ordem importa, flag que falta, caso de borda). Codifique-os aqui.
5. `## Verificação` — como saber que deu certo.

Material de referência grande (specs, exemplos longos) vai em **arquivos separados**
no diretório da skill (`reference.md`, `scripts/`, `templates/`), citados do SKILL.md
para carregar só quando preciso. Use `${CLAUDE_SKILL_DIR}` para referenciar scripts
empacotados em comandos bash.

## `disable-model-invocation`: quando usar

Ligue (`true`) para fluxos com **efeito colateral** ou de timing sensível — deploy,
commit, release, enviar mensagem. Você não quer o Claude disparando sozinho. Para
skills de conhecimento/refactor reversível, deixe desligado (default) para o disparo
automático funcionar.

## Procedimento

1. Confirme o **nome** (kebab-case) e o **escopo** (pessoal vs projeto) com o usuário
   se houver dúvida.
2. Crie `<local>/<nome>/SKILL.md` com o frontmatter e o corpo acima.
3. Escreva a `description` com caso-chave + frases-gatilho.
4. Preencha `## Armadilhas` com os gotchas reais (não invente — extraia do
   procedimento/sessão que originou a skill).
5. Avise que skills sob `~/.claude/skills/` e `.claude/skills/` são detectadas **ao
   vivo** (sem reiniciar), salvo criar um diretório `skills/` de nível superior que
   não existia ao abrir a sessão (aí precisa reiniciar).

## Armadilhas

- **Não duplique skills existentes.** Antes de criar, liste `~/.claude/skills/` e a
  listagem de skills disponíveis; se já existe algo parecido (ex.: há um
  `skill-creator` oficial de plugin), avise o usuário e ofereça estender em vez de clonar.
- O `name` do frontmatter **não** muda o comando (exceto SKILL.md na raiz de plugin)
  — o comando é o nome do diretório.
- `description` fraca = skill nunca dispara. Sem gatilhos concretos, o Claude não
  sabe quando aplicá-la.
- Procedimento que é "sempre antes/depois de X" automaticamente é **hook**
  (settings.json), não skill — o harness executa hooks, o Claude não. Veja a skill
  [update-config].

## Referência

Doc oficial: https://code.claude.com/docs/en/skills (Agent Skills,
https://agentskills.io). Consulte para recursos avançados: subagent execution,
injeção dinâmica de contexto (`${...}`), `arguments`, namespacing de plugin.
