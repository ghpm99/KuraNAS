---
name: quality-gate-fix
description: >-
  Roda `make ci` (ou equivalente), identifica cada categoria de falha (lint,
  format, bandit, tests, mypy) e aplica as correções necessárias até o gate
  passar. TRIGGER quando: usuário disser "quality gate não está passando",
  "make ci falhou", "ci quebrando", "arruma o lint", "fix flake8/bandit/black";
  antes de declarar qualquer tarefa concluída quando o projeto tem `make ci`.
  SKIP: projetos sem Makefile ou sem target ci/lint/test.
---

# Quality Gate Fix

## Passo 1 — Roda o gate e captura a saída completa

```bash
make ci 2>&1 | tee /tmp/ci_output.txt
```

Se não existir `make ci`, procura o equivalente:
```bash
# Descobre o target correto
grep -E "^ci|^lint|^test|^check" Makefile | head -20

# Alternativas comuns
make lint && make test
make check
```

## Passo 2 — Classifica as falhas

Lê `/tmp/ci_output.txt` e mapeia para categorias:

| Saída | Categoria | Fix |
|-------|-----------|-----|
| `black --check` / `would reformat` | Formatação | `black .` |
| `isort --check` / `ERROR` | Import order | `isort .` |
| `flake8 E...` / `W...` | Style lint | Edita os arquivos |
| `bandit` / `Issue: [B...]` | Security lint | Edita ou adiciona `# nosec` justificado |
| `mypy` / `error:` | Tipagem | Adiciona tipos ou cast |
| `FAILED` / `AssertionError` | Testes | Analisa e corrige |
| `coverage` / `FAIL Required test coverage` | Cobertura | Adiciona testes |

## Passo 3 — Aplica correções por categoria

### Formatação (black / isort)
```bash
black .
isort .
```
Confirma que o resultado não quebra nada:
```bash
python -m py_compile <arquivo_alterado>
```

### Flake8 — erros comuns

| Código | Significado | Fix |
|--------|-------------|-----|
| E501 | Linha muito longa | Quebra a linha ou adiciona `# noqa: E501` (último recurso) |
| F401 | Import não usado | Remove o import |
| F811 | Redefinição de import | Remove o duplicado |
| E302/E303 | Linhas em branco | Adiciona/remove linhas conforme a regra |
| W291/W293 | Espaço no final | Remove espaço trailing |
| E711 | Comparação com None | Troca `== None` por `is None` |

```bash
# Ver todos os erros de um arquivo específico
flake8 path/to/file.py

# Auto-fix básico com autopep8 (se disponível)
autopep8 --in-place --select E302,E303,W291,W293 path/to/file.py
```

### Bandit — falsos positivos e correções

```bash
# Ver detalhes do issue
bandit -r app/ -ll

# Falso positivo (justificado): adiciona nosec com explicação
result = subprocess.run(cmd, shell=False)  # nosec B603 — cmd é lista estática

# Issues reais mais comuns:
# B101 assert usado fora de teste → substituir por if/raise
# B105/B106 hardcoded password → mover para settings/env var
# B311 random não-criptográfico → usar secrets.choice() se for segurança
# B608 SQL injection → usar ORM ou parameterização
```

### Mypy — problemas comuns

```python
# Argumento sem tipo → adiciona anotação
def minha_funcao(valor):           # error: Function is missing a type annotation
def minha_funcao(valor: str) -> None:  # ok

# Optional não tratado → adiciona guard
obj = queryset.first()
obj.save()                        # error: Item "None" of "Optional[X]" has no attribute "save"
if obj is not None:
    obj.save()                    # ok

# Dict/List sem parâmetro de tipo
from typing import List, Dict
def foo(items: List[str]) -> Dict[str, int]: ...
```

### Testes falhando

```bash
# Roda só os testes que falharam
python manage.py test app.tests.FailingTestCase

# Com mais verbosidade
python manage.py test app.tests.FailingTestCase -v 2

# Se for problema de memória, usa a skill django-test-memory-debug
# Se for problema de mock, usa a skill django-test-refactor-behavior
```

### Cobertura insuficiente

```bash
# Vê o relatório de cobertura
coverage report --skip-covered | head -40

# Identifica os arquivos mais descobertos
coverage report --skip-covered | sort -k4 -n | head -20

# Roda coverage com html para inspecionar linhas
coverage html && open htmlcov/index.html
```

## Passo 4 — Loop até passar

Depois de cada rodada de correções:
```bash
make ci 2>&1 | tail -30
```

Repete até ver `OK` ou `Passed` sem erros.

## Passo 5 — Confirma que não houve regressão

```bash
# Roda a suite completa uma vez mais
make test

# Verifica git diff para entender o que mudou
git diff --stat
```

## Armadilhas comuns

- **`# noqa` sem código específico** — flake8 aceita mas inibe todos os checks.
  Sempre usa `# noqa: E501` em vez de `# noqa`.
- **`# nosec` sem justificativa** — bandit aceita mas fica difícil de revisar.
  Adiciona comentário explicando por que é falso positivo.
- **`black` conflitando com `flake8 E501`** — se o projeto tem ambos, confere
  se `max-line-length` do flake8 está alinhado com o `line-length` do black
  em `pyproject.toml` ou `setup.cfg`.
- **Testes de integração que precisam de DB Postgres** — podem falhar em
  ambiente sem Postgres. Usa `@skipUnless(connection.vendor == "postgresql", ...)`
  nos testes que dependem de Postgres.
