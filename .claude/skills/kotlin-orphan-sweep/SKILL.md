---
name: kotlin-orphan-sweep
description: "Detecta e remove imports e helpers privados que ficaram sem uso após um refactor em arquivos Kotlin/Java, usando a heurística de ocorrência única e o compilador como rede de segurança. Use quando o usuário disser 'limpa os imports não usados', 'ficou código morto?', 'tira os imports órfãos', 'tem helper sem uso?', ou logo após um refactor grande que trocou wrappers/APIs."
---

# /kotlin-orphan-sweep

Remove imports e funções privadas que ficaram órfãos após um refactor. São
**warnings**, não erros — mas sujam a base. Esta skill limpa com segurança.

## Quando invocar

- "Limpa os imports não usados / órfãos"
- "Ficou código morto depois do refactor?"
- "Tem helper privado sem uso?"
- Logo após migrar telas/trocar wrappers em vários arquivos

## Passo 0 — vale a pena bloquear?

Cheque se o build trata warning como erro:

```
grep -rn "allWarningsAsErrors\|-Werror" app/build.gradle.kts build.gradle.kts gradle.properties
```

Se **não** houver, imports órfãos não quebram o build — é débito de limpeza, não
bloqueio. Diga isso ao usuário e prossiga só se ele quiser arrumar.

## Heurística de detecção (ocorrência única)

Um símbolo importado que aparece **uma única vez** no arquivo só pode estar na
própria linha de `import` → é órfão.

```bash
# para cada arquivo e símbolo suspeito:
if grep -qE "^import .*\.<Symbol>$" "$f" && [ "$(grep -cE "\b<Symbol>\b" "$f")" -eq 1 ]; then echo "órfão: $f <Symbol>"; fi
```

Cuidados na contagem:
- Use `\b<Symbol>\b` (limite de palavra) — assim `Box` **não** casa dentro de
  `CenterBox`/`SnackbarHost`.
- Símbolos comuns (`Box`, `Column`, `Row`, `Alignment`, `Arrangement`) geralmente
  têm contagem > 1 (usados no corpo). Só remova os que derem exatamente 1.
- Símbolos típicos de sobra após migração de scaffold/estado:
  `CircularProgressIndicator`, `Button`, `Scaffold`, `TopAppBar`, ícones nav-only.

## Remoção

Remova a **linha de import exata** (não substring) para não apagar import vizinho:

```bash
sed -i -E "/^import [a-z0-9.]+\.<Symbol>$/d" "$f"
```

**Helpers privados órfãos**: se `private fun X` aparece só na definição
(`grep -c "X" == 1`), remova a função inteira com a ferramenta de edição (não sed,
para casar o corpo multi-linha com precisão).

## Rede de segurança

**O compilador é a verificação final.** Remover um import que ainda está em uso vira
`unresolved reference` (erro), não silêncio — então a estratégia "remover agressivo +
compilar" é segura. Compile depois da varredura (veja [android-verify]); se algo
quebrar, foi remoção indevida — reponha aquele import.

## Re-checagem

Após remover, rode a heurística de novo para confirmar zero órfãos restantes dos
símbolos visados.
