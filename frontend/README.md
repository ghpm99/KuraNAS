<!-- Improved compatibility of back to top link: See: https://github.com/othneildrew/Best-README-Template/pull/73 -->
<a id="readme-top"></a>
<!--
*** Thanks for checking out the Best-README-Template. If you have a suggestion
*** that would make this better, please fork the repo and create a pull request
*** or simply open an issue with the tag "enhancement".
*** Don't forget to give the project a star!
*** Thanks again! Now go create something AMAZING! :D
-->



<!-- PROJECT SHIELDS -->
<!--
*** I'm using markdown "reference style" links for readability.
*** Reference links are enclosed in brackets [ ] instead of parentheses ( ).
*** See the bottom of this document for the declaration of the reference variables
*** for contributors-url, forks-url, etc. This is an optional, concise syntax you may use.
*** https://www.markdownguide.org/basic-syntax/#reference-style-links
-->
[![Contributors][contributors-shield]][contributors-url]
[![Forks][forks-shield]][forks-url]
[![Stargazers][stars-shield]][stars-url]
[![Issues][issues-shield]][issues-url]
[![project_license][license-shield]][license-url]
[![LinkedIn][linkedin-shield]][linkedin-url]



<!-- PROJECT LOGO -->
<br />
<div align="center">
  <a href="https://github.com/ghpm99/KuraNAS">
    <img src="images/logo.png" alt="Logo" width="80" height="80">
  </a>

<h3 align="center">KuraNAS Frontend</h3>

  <p align="center">
    Aplicação web do KuraNAS construída com React + TypeScript + Vite, com estrutura feature-first.
    <br />
    <a href="https://github.com/ghpm99/KuraNAS"><strong>Explore the docs »</strong></a>
    <br />
    <br />
    <a href="https://github.com/ghpm99/KuraNAS">View Demo</a>
    &middot;
    <a href="https://github.com/ghpm99/KuraNAS/issues/new?labels=bug&template=bug-report---.md">Report Bug</a>
    &middot;
    <a href="https://github.com/ghpm99/KuraNAS/issues/new?labels=enhancement&template=feature-request---.md">Request Feature</a>
  </p>
</div>



<!-- TABLE OF CONTENTS -->
<details>
  <summary>Table of Contents</summary>
  <ol>
    <li>
      <a href="#about-the-project">About The Project</a>
      <ul>
        <li><a href="#built-with">Built With</a></li>
      </ul>
    </li>
    <li>
      <a href="#getting-started">Getting Started</a>
      <ul>
        <li><a href="#prerequisites">Prerequisites</a></li>
        <li><a href="#installation">Installation</a></li>
      </ul>
    </li>
    <li><a href="#usage">Usage</a></li>
    <li><a href="#roadmap">Roadmap</a></li>
    <li><a href="#contributing">Contributing</a></li>
    <li><a href="#license">License</a></li>
    <li><a href="#contact">Contact</a></li>
    <li><a href="#acknowledgments">Acknowledgments</a></li>
  </ol>
</details>



<!-- ABOUT THE PROJECT -->
## About The Project

SPA do KuraNAS construída com React 19, TypeScript e Vite. Consome a API REST do backend
(`/api/v1`) e adota estrutura **feature-first** para os domínios críticos (`files`, `music`,
`videos`), com lógica e chamadas HTTP em hooks/providers — não em componentes de render.

<p align="right">(<a href="#readme-top">back to top</a>)</p>



### Built With

* [![React][React.js]][React-url]
* [![TypeScript][TypeScript.org]][TypeScript-url]
* [![Vite][Vite.dev]][Vite-url]
* [![MUI][MUI.com]][MUI-url]
* [![React Query][ReactQuery.dev]][ReactQuery-url]

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- GETTING STARTED -->
## Getting Started

Como rodar a aplicação web localmente.

### Prerequisites

* Node.js 20+
* Yarn 1.x

### Installation

1. Clone o repositório
   ```sh
   git clone https://github.com/ghpm99/KuraNAS.git
   ```
2. Instale as dependências
   ```sh
   cd frontend && yarn
   ```
3. Configure o `.env` (ou `.env.development`) apontando para o backend
   ```dotenv
   VITE_API_URL=http://localhost:8000
   ```
4. Inicie o dev server
   ```sh
   yarn dev
   ```

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- USAGE EXAMPLES -->
## Usage

### Estrutura

```text
frontend/
├── src/
│   ├── app/            # composição de rotas e inicialização da aplicação
│   ├── components/     # composição de tela, layout e domínios não migrados
│   ├── features/       # ownership por domínio (files, music, videos)
│   ├── pages/          # wrappers finos de rota
│   ├── service/        # clientes e serviços de API
│   ├── shared/         # utilitários compartilhados cross-feature
│   ├── types/          # tipos compartilhados
│   └── utils/          # utilitários
├── public/
├── jest.config.js
├── eslint.config.js
└── vite.config.ts
```

### Scripts

| Comando | Descrição |
| --- | --- |
| `yarn dev` | Dev server (Vite) |
| `yarn build` | Build de produção |
| `yarn preview` | Preview local do build |
| `yarn lint` | ESLint |
| `yarn test --watchAll=false` | Testes (Jest) |
| `yarn test:watch` | Testes em watch |
| `yarn coverage` | Cobertura |
| `yarn typecheck:test` | Typecheck da config de testes |
| `yarn format` | Format |

### Variáveis de Ambiente

Variável suportada:

- `VITE_API_URL`: URL base da API (sem `/api/v1` no final).

Comportamento da URL base:

- Se `globalThis.__KURANAS_API_URL__` existir em runtime, ela tem prioridade.
- Senão, usa `VITE_API_URL`.
- Sem variável, fallback para caminho relativo (`/api/v1`).

### API e i18n

- O frontend consome a API via `src/service/index.ts` usando `getApiV1BaseUrl()` (`src/service/apiUrl.ts`).
- Textos visíveis devem vir de tradução via `useI18n()`.
- Não adicionar texto hardcoded em componentes.
- Novas mensagens devem ser adicionadas primeiro em `backend/translations` e consumidas por chave.

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- ROADMAP -->
## Roadmap

- [ ] Migrar domínios restantes para `src/features/*`
- [ ] Elevar cobertura acima dos mínimos globais
- [ ] Componentização compartilhada de estados (loading/erro/vazio)

See the [open issues](https://github.com/ghpm99/KuraNAS/issues) for a full list of proposed features (and known issues).

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- CONTRIBUTING -->
## Contributing

Contributions are what make the open source community such an amazing place to learn, inspire, and create. Any contributions you make are **greatly appreciated**.

Antes de alterar código frontend, siga `docs/standards/frontend-standards.md`. Pontos obrigatórios:
domínios `files/music/videos` evoluem prioritariamente em `src/features/*`; lógica e chamadas HTTP
em hooks/providers; uso do alias `@/...`; cobertura mínima global em `jest.config.js`
(90% lines/functions/statements e 89% branches).

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feat/AmazingFeature`)
3. Commit your Changes (`git commit -m 'feat: add some AmazingFeature'`)
4. Push to the Branch (`git push origin feat/AmazingFeature`)
5. Open a Pull Request

<p align="right">(<a href="#readme-top">back to top</a>)</p>

### Top contributors:

<a href="https://github.com/ghpm99/KuraNAS/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=ghpm99/KuraNAS" alt="contrib.rocks image" />
</a>



<!-- LICENSE -->
## License

Nenhum arquivo de licença é distribuído com o projeto até o momento. Defina uma licença em `LICENSE.txt` na raiz do repositório.

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- CONTACT -->
## Contact

ghpm99 - ghpm99@gmail.com

Project Link: [https://github.com/ghpm99/KuraNAS](https://github.com/ghpm99/KuraNAS)

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- ACKNOWLEDGMENTS -->
## Acknowledgments

* [Vite](https://vite.dev/)
* [MUI](https://mui.com/)
* [TanStack Query](https://tanstack.com/query/latest)
* [Best-README-Template](https://github.com/othneildrew/Best-README-Template)
* [Img Shields](https://shields.io)

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- MARKDOWN LINKS & IMAGES -->
<!-- https://www.markdownguide.org/basic-syntax/#reference-style-links -->
[contributors-shield]: https://img.shields.io/github/contributors/ghpm99/KuraNAS.svg?style=for-the-badge
[contributors-url]: https://github.com/ghpm99/KuraNAS/graphs/contributors
[forks-shield]: https://img.shields.io/github/forks/ghpm99/KuraNAS.svg?style=for-the-badge
[forks-url]: https://github.com/ghpm99/KuraNAS/network/members
[stars-shield]: https://img.shields.io/github/stars/ghpm99/KuraNAS.svg?style=for-the-badge
[stars-url]: https://github.com/ghpm99/KuraNAS/stargazers
[issues-shield]: https://img.shields.io/github/issues/ghpm99/KuraNAS.svg?style=for-the-badge
[issues-url]: https://github.com/ghpm99/KuraNAS/issues
[license-shield]: https://img.shields.io/github/license/ghpm99/KuraNAS.svg?style=for-the-badge
[license-url]: https://github.com/ghpm99/KuraNAS/blob/master/LICENSE.txt
[linkedin-shield]: https://img.shields.io/badge/-LinkedIn-black.svg?style=for-the-badge&logo=linkedin&colorB=555
[linkedin-url]: https://linkedin.com/in/linkedin_username
[product-screenshot]: images/screenshot.png
<!-- Shields.io badges. You can a comprehensive list with many more badges at: https://github.com/inttter/md-badges -->
[React.js]: https://img.shields.io/badge/React-20232A?style=for-the-badge&logo=react&logoColor=61DAFB
[React-url]: https://react.dev/
[TypeScript.org]: https://img.shields.io/badge/TypeScript-3178C6?style=for-the-badge&logo=typescript&logoColor=white
[TypeScript-url]: https://www.typescriptlang.org/
[Vite.dev]: https://img.shields.io/badge/Vite-646CFF?style=for-the-badge&logo=vite&logoColor=white
[Vite-url]: https://vite.dev/
[MUI.com]: https://img.shields.io/badge/MUI-007FFF?style=for-the-badge&logo=mui&logoColor=white
[MUI-url]: https://mui.com/
[ReactQuery.dev]: https://img.shields.io/badge/React_Query-FF4154?style=for-the-badge&logo=reactquery&logoColor=white
[ReactQuery-url]: https://tanstack.com/query/latest
