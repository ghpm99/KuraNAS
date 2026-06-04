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

<h3 align="center">KuraNAS Android</h3>

  <p align="center">
    App Android moderno do KuraNAS em Kotlin + Jetpack Compose, com Media3 e arquitetura por feature.
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

App Android moderno do KuraNAS, escrito em **Kotlin** com **Jetpack Compose** e
**Material 3**. Consome a API REST do backend (`/api/v1`) e inclui reprodução de mídia
via **Media3 (ExoPlayer)**, incluindo HLS.

Diferente do módulo `mobile/` (legado, Java/XML para API 16), este módulo (`android/`)
tem `minSdk 33`, `targetSdk 36` e `applicationId` `com.kuranas.android`.

Domínios cobertos por feature: `connection`, `diary`, `files`, `home`, `images`, `jobs`,
`music`, `notifications`, `search`, `settings`, `video`.

<p align="right">(<a href="#readme-top">back to top</a>)</p>



### Built With

* [![Kotlin][Kotlin.com]][Kotlin-url]
* [![Jetpack Compose][Compose.dev]][Compose-url]
* [![Hilt][Hilt.dev]][Hilt-url]
* [![Retrofit][Retrofit.com]][Retrofit-url]
* [![Media3][Media3.dev]][Media3-url]

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- GETTING STARTED -->
## Getting Started

Como compilar e instalar o app Android moderno.

### Prerequisites

* JDK 17
* Android SDK com `compileSdk 36` instalado
* `local.properties` com `sdk.dir` apontando para o Android SDK
* Dispositivo/emulador com Android 13+ (`minSdk 33`)

### Installation

1. Clone o repositório
   ```sh
   git clone https://github.com/ghpm99/KuraNAS.git
   ```
2. Gere o APK de debug
   ```sh
   cd android && ./gradlew assembleDebug
   ```
3. Instale em um dispositivo conectado
   ```sh
   cd android && ./gradlew installDebug
   ```

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- USAGE EXAMPLES -->
## Usage

### Estrutura

```text
android/
├── app/
│   └── src/main/java/com/kuranas/android/
│       ├── core/           # infraestrutura compartilhada (rede, DI, base)
│       ├── feature/        # ownership por domínio (files, music, video, images, ...)
│       ├── navigation/     # grafo de navegação Compose
│       └── ui/             # tema, componentes e design system
├── gradle/
│   └── libs.versions.toml  # version catalog
├── build.gradle.kts
├── settings.gradle.kts
└── gradle.properties
```

### Stack técnico

| Área | Tecnologia |
| --- | --- |
| Linguagem | Kotlin |
| UI | Jetpack Compose + Material 3 |
| DI | Hilt |
| Navegação | Navigation Compose |
| HTTP | Retrofit + kotlinx.serialization |
| Concorrência | Kotlin Coroutines |
| Persistência local | DataStore Preferences + Security Crypto |
| Imagens | Coil |
| Mídia | Media3 (ExoPlayer, UI, Session, HLS) |

### Comandos

| Comando | Descrição |
| --- | --- |
| `./gradlew assembleDebug` | Build debug |
| `./gradlew assembleRelease` | Build release |
| `./gradlew installDebug` | Instala o APK debug no dispositivo conectado |
| `./gradlew test` | Testes unitários |
| `./gradlew connectedAndroidTest` | Testes instrumentados |

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- ROADMAP -->
## Roadmap

- [ ] Fila/mini-player e reprodução em segundo plano (Música)
- [ ] Ampliar cobertura de testes por feature
- [ ] Refinar design system (glassmorphism navy/blue)

See the [open issues](https://github.com/ghpm99/KuraNAS/issues) for a full list of proposed features (and known issues).

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- CONTRIBUTING -->
## Contributing

Contributions are what make the open source community such an amazing place to learn, inspire, and create. Any contributions you make are **greatly appreciated**.

Mantenha a arquitetura por feature (`feature/<domain>`), Compose + Material 3, e injeção via Hilt.
Rode `./gradlew test` e `./gradlew assembleDebug` antes de abrir o PR.

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

* [Jetpack Compose](https://developer.android.com/jetpack/compose)
* [Media3](https://developer.android.com/media/media3)
* [Hilt](https://dagger.dev/hilt/)
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
[Kotlin.com]: https://img.shields.io/badge/Kotlin-7F52FF?style=for-the-badge&logo=kotlin&logoColor=white
[Kotlin-url]: https://kotlinlang.org/
[Compose.dev]: https://img.shields.io/badge/Jetpack_Compose-4285F4?style=for-the-badge&logo=jetpackcompose&logoColor=white
[Compose-url]: https://developer.android.com/jetpack/compose
[Hilt.dev]: https://img.shields.io/badge/Hilt-2196F3?style=for-the-badge&logo=android&logoColor=white
[Hilt-url]: https://dagger.dev/hilt/
[Retrofit.com]: https://img.shields.io/badge/Retrofit-48B983?style=for-the-badge&logo=square&logoColor=white
[Retrofit-url]: https://square.github.io/retrofit/
[Media3.dev]: https://img.shields.io/badge/Media3-3DDC84?style=for-the-badge&logo=android&logoColor=white
[Media3-url]: https://developer.android.com/media/media3
