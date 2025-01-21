# pipo-dispatcher
[![License](https://img.shields.io/github/license/arraial/pipo-dispatcher)](https://opensource.org/licenses/MIT)
[![Build](https://github.com/arraial/pipo-dispatcher/actions/workflows/docker.yml/badge.svg)](https://github.com/arraial/pipo-dispatcher/actions/workflows/docker.yml)
[![Version](https://img.shields.io/github/v/tag/arraial/pipo-dispatcher)](https://github.com/arraial/pipo-dispatcher/releases)
[![Docker Image](https://img.shields.io/docker/image-size/arraial/pipo_dispatcher/latest)](https://hub.docker.com/r/arraial/pipo_dispatcher)

Pipo service handling hub request forwarding to services

## Installation

### Runtime prerequisites
The application is compatible with Windows and Linux based systems.
[Docker](https://docs.docker.com/engine/install/) + [Docker Compose](https://docs.docker.com/compose/install/) are assumed to be installed and configured.

### Development
One may leverage VS Code Devcontainer for a simplified setup or other suitable option, as describbed in [Manual Setup](#manual).

#### Visual Studio Devcontainer
Devcontainer functionality can be used by choosing option `Dev Containers: Open Folder in Container...` in VS Code.

#### Manual Setup
For these guiding steps a [compatible version](go.mod) of Golang is assumed to be installed.

To setup the development environment and being able to run the test suite do:
```bash
make dev_setup
```

Build the app container image with:
```bash
make image
```

For additional help try:
```bash
make help
```

## How to run

### Test suite
```bash
make test
```

### Containerized application
Before running the following command make sure `.env` was created and filled based on the available [example](.env.example).

Start the container with
```bash
make run_image
```

## License
This project is licensed under the MIT License - see [LICENSE](LICENSE) file for details.
