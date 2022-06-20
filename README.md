# previewd

[![License](https://img.shields.io/github/license/clarkezone/previewd.svg)](https://github.com/clarkezone/previewd/blob/main/LICENSE) [![Go Report Card](https://goreportcard.com/badge/github.com/clarkezone/previewd)](https://goreportcard.com/report/github.com/clarkezone/previewd) [![Build and Tests](https://github.com/clarkezone/previewd/workflows/run%20tests/badge.svg)](https://github.com/clarkezone/previewd/actions?query=workflow%3A%22run+tests%22) [![Coverage Status](https://coveralls.io/repos/github/clarkezone/previewd/badge.svg?branch=main)](https://coveralls.io/github/clarkezone/previewd?branch=main) [![Go Reference](https://pkg.go.dev/badge/github.com/clarkezone/previewd.svg)](https://pkg.go.dev/github.com/clarkezone/previewd)

## Description

`previewd` is a daemon that is primarily designed to be deployed into a kubernetes cluster to facilitate previewing and hosting of static websites built using a static site generator such as Jekyll, Hugo or Publish.

```mermaid
graph  LR
  client([browser])-..->ingress[Ingress];
  webhook([git webhook])-..->hookingress[webhook receiver<BR>ingress];
  ingress-->service[Service];
  hookingress-->previewdservice
  repo-->webhook
  subgraph git
  repo
  end
  subgraph cluster
  hookingress;
  ingress[Web<BR>frontend<BR>ingress];
  previewdservice-->pod3[previewd<BR>jobmanager pod]
  repo-..pull content..->pod3
  pod3-->renderjob
  renderjob[RenderJob<BR>Jekyll image]
  pod3-->source
  source-->renderjob
  renderjob-->render
  pod1-->render
  pod2-->render
  source[Source<BR>volume]
  render[Render<BR>volume]
  service-->pod1[nginx replica 1];
  service-->pod2[inginx replica 2];

  end
  classDef plain fill:#ddd,stroke:#fff,stroke-width:4px,color:#000;
  classDef k8s fill:#326ce5,stroke:#fff,stroke-width:4px,color:#fff;
  classDef cluster fill:#fff,stroke:#bbb,stroke-width:2px,color:#326ce5;
  classDef volume fill:#fff,stroke:#bbb,stroke-width:2px,color:#326ce5;
  class ingress,hookingress,service,pod1,pod2,pod3,previewdservice, k8s;
  class client plain;
  class cluster cluster;
  class source volume
  class render volume
```

## Installation and use

This project is still in development and as such we don't yet have instructions for how to use. That said you can build the code and run the tests. The backlog is maintained in [docs/workbacklog.md](docs/workbacklog.md)

### Install Tools

1. Install a recent version of golang, we recommend 1.17 or greater. [https://go.dev/doc/install](https://go.dev/doc/install)
2. ensure that your path includes: `export PATH=$PATH:$HOME/go/bin` so that tools installed with `go install` work correctly
3. Install `make` (debian linux: `sudo apt install make`)
4. Install `gcc` (debian linux: `sudo apt install build-essential`)
5. Install [`pre-commit`](https://pre-commit.com/) (debian linux: `sudo apt install precommit`)
6. If you are planning on submitting a PR to this repo, install the git pre-commit hook (`pre-commit install`)
7. Install [`shellcheck`](https://github.com/koalaman/shellcheck) (`sudo apt install shellcheck`)
8. Install tools other golang based linting tools `make install-tools`
9. Install [`k3s`](https://github.com/k3s-io/k3s) (`curl -sfL https://get.k3s.io | sh -s - --write-kubeconfig-mode 644`)
10. If you are planning to use VSCode, ensure you have all of the golang tooling installed

### Dev Setup

1. Configure environment variables in shell. There is a .env file in the scripts directory to establish environment variables needed for unit tests. These need to be applied in your shell and in vscode. You can add the following to your .bashrc or manually.

   ```bash
   export $(cat scripts/.previewd_test.env | xargs)
   ```

2. (optional) if you are planning to debug tests in vscode, you'll need to tell VScode about the environment variables and also enable integration tests. In VS code, go to the command palette and search for `Preferences: Open Remote settings (JSON)` and add the following snippet:

   ```json
   {
     "go.buildFlags": ["-tags=unit,integration"],
     "go.buildTags": "-tags=unit,integration",
     "go.testTags": "-tags=unit,integration"
   }
   ```

   Then search for `User Settings` and add the following snippet for environment variables:

   ```json
   "go.testEnvFile": "/home/james/.previewd_test.env",
   ```

3. Edit `internal/testutils.go` and change the value returned by `GetTestConfigPath()` to point to a valid kubeconfig for your test cluster (eg `~/.kube.config` in typical setups or `/etc/rancher/k3s/k3s.yaml` if you followed the instructions above and installed k3s)

### Build from Source

To compile the main previewd executable:

```bash
make build
```

Which will produce a binary at <projectroot>/bin/previewd

To build the docker container

```bash
docker build .
```

### Run tests

This project comes with both unit and integration tests. Unit tests run standalone with no dependencies other than test environment configuration. Integration tests require a kubernetes cluster.

For unit tests, use:

```bash
make test
```

For integration test, since these depend on having a valid k8s cluster to work properly, make sure you followed step 3 in the dev setup list above. Then use:

```bash
make integration
```

TODO: run tests in VS (exports)
