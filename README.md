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

<h3 align="center">KuraNAS</h3>

  <p align="center">
    Sistema NAS pessoal com backend em Go e frontend em React para gerenciamento de arquivos, mídia e organização.
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

KuraNAS é um sistema NAS (Network Attached Storage) pessoal com backend em Go e frontend em
React/TypeScript. O backend serve o frontend como SPA e expõe uma API REST em `/api/v1`.

O repositório é um monorepo com os seguintes módulos:

- **Backend** (`backend/`): API HTTP, regras de negócio, workers e i18n.
- **Frontend** (`frontend/`): SPA React + TypeScript, estrutura `feature-first` para domínios críticos (`files`, `music`, `videos`).
- **Mobile** (`mobile/`): app Android nativo (API 16) em Java + XML + AppCompat, com ownership incremental por feature.
- **Plugin** (`plugin/`): extensão Chrome MV3 modularizada (`src/background`, `src/shared`) para captura de mídia.
- **Build integrado**: empacotamento final em `build/`.

### Estrutura

```text
.
├── backend/            # API, workers, banco, i18n e scripts
├── frontend/           # Aplicação web (Vite + React + TypeScript)
├── mobile/             # App Android (API 16, Java + XML + AppCompat)
├── plugin/             # Extensão Chrome (Manifest V3)
├── docs/               # Padrões de engenharia e documentação funcional
├── build/              # Saída do build integrado (gerado)
└── Makefile            # Pipeline local de build/qualidade
```

<p align="right">(<a href="#readme-top">back to top</a>)</p>



### Built With

* [![Go][Go.dev]][Go-url]
* [![Gin][Gin.com]][Gin-url]
* [![PostgreSQL][Postgres.org]][Postgres-url]
* [![React][React.js]][React-url]
* [![TypeScript][TypeScript.org]][TypeScript-url]
* [![Vite][Vite.dev]][Vite-url]

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- GETTING STARTED -->
## Getting Started

Setup rápido para desenvolvimento local (backend + frontend).

### Prerequisites

* Go 1.24+
* Node.js 20+
* npm 10+
* Yarn 1.x
* Make
* JDK 17+
* Android SDK + Build Tools para `compileSdk 35` (apenas para o módulo mobile)

### Installation

1. Clone o repositório
   ```sh
   git clone https://github.com/ghpm99/KuraNAS.git
   ```
2. Instale as dependências do frontend
   ```sh
   cd frontend && yarn
   ```
3. Configure as variáveis do backend em `backend/.env` (detalhes em [`backend/README.md`](backend/README.md))
4. Inicie o backend (modo `dev`, porta `8000`)
   ```sh
   make -C backend run
   ```
5. Em outro terminal, inicie o frontend
   ```sh
   cd frontend && yarn dev
   ```

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- USAGE EXAMPLES -->
## Usage

### Build Integrado

Gera frontend + backend e organiza artefatos em `build/`:

```bash
make
```

Limpeza:

```bash
make clean
```

### Onboarding por Stack

Backend:

```bash
cd backend && go test ./... -cover
make -C backend run
```

Frontend:

```bash
cd frontend && yarn lint
cd frontend && yarn test --watchAll=false
cd frontend && yarn build
```

Mobile:

```bash
cd mobile && ./gradlew test
cd mobile && ./gradlew assembleDebug
```

Plugin:

```bash
cd plugin && npm ci
cd plugin && npm run lint
cd plugin && npm test
```

Pipeline local completa:

```bash
make ci
```

### Backup e segunda cópia

O KuraNAS executa um backup incremental das raízes de armazenamento para a pasta de destino configurada na UI (Configurações → Backup), com retenção de versões em `_versions/`. A **segunda cópia** (HD externo de 2 TB, idealmente desconectável) fica fora do aplicativo: sincronize o diretório de backup para o HD externo pelo próprio SO — por exemplo, `robocopy <destino-do-backup> <hd-externo> /MIR` agendado pelo Agendador de Tarefas do Windows. O sistema não gerencia mídia desconectável.

### Internacionalização

- Não hardcode texto visível para usuário.
- Backend e frontend devem usar as mesmas chaves em `backend/translations`.
- O frontend obtém traduções via endpoint de configuração do backend.

### Documentação por Módulo

- [README do backend](backend/README.md)
- [README do frontend](frontend/README.md)
- [README do mobile](mobile/README.md)
- [README do plugin](plugin/README.md)
- [Padrão backend](docs/standards/backend-standards.md)
- [Padrão frontend](docs/standards/frontend-standards.md)
- [Padrão mobile](docs/standards/mobile-standards.md)
- [Padrão plugin](docs/standards/plugin-standards.md)

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- ROADMAP -->
## Roadmap

- [ ] Pipeline de ingestão de mídia (Library Paths, importadores, watch)
- [ ] Documentação OpenAPI da API `/api/v1`
- [ ] Elevar cobertura de testes nas stacks

See the [open issues](https://github.com/ghpm99/KuraNAS/issues) for a full list of proposed features (and known issues).

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- CONTRIBUTING -->
## Contributing

Contributions are what make the open source community such an amazing place to learn, inspire, and create. Any contributions you make are **greatly appreciated**.

Consulte o README e os padrões do módulo correspondente em `docs/standards/` antes de alterar código.

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

* [Go](https://go.dev/)
* [React](https://react.dev/)
* [Vite](https://vite.dev/)
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
[Go.dev]: https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white
[Go-url]: https://go.dev/
[Gin.com]: https://img.shields.io/badge/Gin-008ECF?style=for-the-badge&logo=gin&logoColor=white
[Gin-url]: https://gin-gonic.com/
[Postgres.org]: https://img.shields.io/badge/PostgreSQL-4169E1?style=for-the-badge&logo=postgresql&logoColor=white
[Postgres-url]: https://www.postgresql.org/
[React.js]: https://img.shields.io/badge/React-20232A?style=for-the-badge&logo=react&logoColor=61DAFB
[React-url]: https://react.dev/
[TypeScript.org]: https://img.shields.io/badge/TypeScript-3178C6?style=for-the-badge&logo=typescript&logoColor=white
[TypeScript-url]: https://www.typescriptlang.org/
[Vite.dev]: https://img.shields.io/badge/Vite-646CFF?style=for-the-badge&logo=vite&logoColor=white
[Vite-url]: https://vite.dev/
