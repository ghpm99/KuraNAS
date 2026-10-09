---
name: compose-screen-standardize
description: "Migra telas Jetpack Compose de Scaffold/TopAppBar montados à mão para o scaffold/top bar compartilhado do design system e unifica os estados loading/erro/vazio em componentes comuns. Use quando o usuário disser 'padroniza os cabeçalhos', 'migra essas telas pro KaworiScaffold/scaffold comum', 'unifica os estados de loading/erro/vazio', 'tira o Scaffold/TopAppBar duplicado de cada tela', ou ao adotar/expandir um design system em telas Compose existentes."
---

# /compose-screen-standardize

Padroniza telas Jetpack Compose substituindo `Scaffold { topBar = TopAppBar(...) }`
montado à mão pelo scaffold compartilhado do design system, e troca os blocos
ad-hoc de loading/erro/vazio pelos componentes comuns. É refactor presentacional:
não muda comportamento, só consolida aparência e remove duplicação.

## Quando invocar

- "Padroniza os cabeçalhos / migra pro scaffold comum"
- "Unifica os estados de loading/erro/vazio"
- "Tira o Scaffold/TopAppBar duplicado de cada tela"
- Ao adotar um design system e querer aplicá-lo às telas já existentes

## Pré-requisitos

1. Confirme que existe (ou crie antes) o scaffold compartilhado e os componentes
   de estado. No kawori-mobile: `core/ui/components/` com `KaworiScaffold`,
   `KaworiTopAppBar`, `LoadingState`, `ErrorState`, `EmptyState`. Se não existir,
   pare e construa a fundação primeiro — sem ela o refactor só espalha duplicação.
2. Identifique as telas-alvo: `grep -rl "TopAppBar" <feature dir>`.

## Procedimento por tela

1. **Trocar o wrapper.** Substitua

   ```kotlin
   Scaffold(
       topBar = { TopAppBar(title = { Text("X") }, navigationIcon = { IconButton(onClick = onNavigateBack) { Icon(ArrowBack, ...) } }) },
       [snackbarHost = { SnackbarHost(snackbar) },]
       [floatingActionButton = { ... },]
   ) { padding -> <corpo> }
   ```

   por

   ```kotlin
   KaworiScaffold(
       title = "X",
       onNavigateBack = onNavigateBack,
       [snackbarHost = { SnackbarHost(snackbar) },]
       [floatingActionButton = { ... },]
   ) { padding -> <corpo> }
   ```

2. **Unificar estados** dentro do `when`:
   - `Box{ CircularProgressIndicator() }` → `LoadingState(Modifier.padding(padding))`
   - `Box{ Column{ Text(erro); Button("Tentar novamente") } }` → `ErrorState(msg, Modifier.padding(padding), onRetry = viewModel::load)`
   - `Box{ Text("Nenhum...") }` (estado de tela cheia) → `EmptyState(title = "Nenhum ...", modifier = Modifier.padding(padding))`

3. **Limpar imports** com [kotlin-orphan-sweep] (ou manualmente): remova `Scaffold`,
   `TopAppBar` e os ícones de navegação que sobraram.

## Armadilhas (o valor real desta skill)

- **NÃO remova `@OptIn(ExperimentalMaterial3Api::class)`** se a tela ainda usa outra
  API experimental (`FilterChip`, `DropdownMenu`/`ExposedDropdownMenuBox`). Só o
  `TopAppBar` saiu — confirme com `grep -c "FilterChip\|ExposedDropdown" <arquivo>`.
- **`snackbarHost` precisa ser repassado** ao scaffold compartilhado. Se o scaffold
  comum não tiver esse parâmetro, adicione-o antes (um `@Composable () -> Unit = {}`).
- **NÃO remova o import de `Icon`/`Icons`** se forem usados no corpo (FAB com
  `Icons.Filled.Add`, ícone de dropdown). Só os nav-only (`ArrowBack`, `IconButton`
  do `navigationIcon`). Cheque a contagem: `grep -c "Icon(" <arquivo>` == 1 ⇒ era só o nav.
- **Balanceamento de chaves** ao envolver um `Column(modifier...)` numa nova lambda
  de scaffold (telas top-level sem cabeçalho): você adiciona um nível
  (`KaworiScaffold(...) { padding -> Column(Modifier.padding(padding)...) { ... } }`)
  e precisa de **uma chave de fechamento a mais** antes do `}` da função. Releia o
  fim da função e confira.
- **Telas top-level dentro do Scaffold da bottom-nav**: aninhar um scaffold é ok; o
  `modifier` recebido já carrega o inset da barra. Passe `modifier` ao
  `KaworiScaffold` e use `Modifier.padding(padding)` no conteúdo.
- **Estados inline contextuais** ("sem pagamentos nesta fatura", passos de wizard)
  NÃO são estados de tela cheia — deixe como `Text` inline, não troque por `EmptyState`.
- **Telas de auth** (login/signup) costumam ter identidade visual própria (barra
  transparente sobre fundo da marca) — deixe de fora salvo pedido explícito.
- **Linhas de lista sob medida** (com valores/cores/tags) NÃO devem virar um
  `ListItem` genérico — perderiam informação. O `KaworiCard` é a base comum delas.

## Verificação

Compile **a cada lote** (não tudo de uma vez) — veja [android-verify]. O compilador
pega remoção de import em uso (vira erro, não silêncio) e chaves desbalanceadas.
Rode também os testes unitários ao final. Atualize `VERSION.md`/roadmap conforme a
convenção do projeto.
