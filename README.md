# chart-verifier

`chart-verifier` is a tool that validates a Helm chart against a configurable list of checks; individual checks can be
included or excluded through command line options. The default set of tests covers Red Hat’s recommendations.

Each check is independent and execution order is not guaranteed. Input is provided through options in
the command line interface; currently the only input is the required `uri` option.

The following checks have been implemented:

| Name | Description
|---|---
| `is-helm-v3` | Checks whether the given `uri` is a Helm v3 chart.
| `has-readme` | Checks whether the Helm chart contains a `README.md` file.
| `contains-test` | Checks whether the Helm chart contains at least one test file.
| `has-minkubeversion` | Checks whether the Helm chart's `Chart.yaml` includes the `minKubeVersion` field.
| `readme-contains-values-schema` | Checks whether the Helm chart `README.md` file contains a `values` schema section.
| `not-contains-crds` | Check whether the Helm chart does not include CRDs.

The following checks are being implemented and/or considered:

| Name | Description
|---|---
| `keywords-are-openshift-categories` | Checks whether the Helm chart's `Chart.yaml` file includes keywords mapped to OpenShift categories.
| `is-commercial-chart` | Checks whether the Helm chart is a Commercial chart.
| `is-community-chart` | Checks whether the Helm chart is a Community chart.
| `not-contains-infra-plugins-and-drivers` | Check whether the Helm chart does not include infra plugins and drivers (network, storage, hardware, etc)
| `can-be-installed-without-manual-prerequisites` |
| `can-be-installed-without-cluster-admin-privileges` |

## Architecture

This tool inspects a Helm
chart URI (`file://`, `https?://`, etc)
and returns either a *positive* result indicating the Helm chart has passed all checks, or a *negative* result indicating
which checks have failed and remedial actions.

The application is separated in two pieces: a command line interface and a library. This is handy because the command
line interface is specific to the user interface, and the library can be generic enough to be used to, for example,
inspect Helm chart bytes in flight.

One positive aspect of the command line interface specificity is that its output can be tailored to the methods of
consumption the user expects; in other words, the command line interface can be programmed in such way it can be
represented as either *YAML* or *JSON* formats, in addition to a descriptive representation tailored to human actors.

Primitive functions to manipulate the Helm chart should be provided, since most checks involve inspecting the contents
of the chart itself; for example, whether a `README.md` file exists, or whether `README.md` contains the `values`'
specification, implicating in offering a cache API layer is required to avoid downloading and unpacking the charts for
each test.

## Getting chart-verifier

Container images built from the source code are hosted in https://quay.io/repository/redhat-certification/chart-verifier
; to download using `docker` execute the following command:

```text
docker pull quay.io/redhat-certification/chart-verifier
```

## Building chart-verifier

To build `chart-verifier` locally, please execute `hack/build.sh` or its PowerShell alternative.

To build `chart-verifier` container image, please execute `hack/build-image.sh` or its PowerShell alternative:

```text
PS C:\Users\igors\GolandProjects\chart-verifier> .\hack\build-image.ps1
[+] Building 15.1s (15/15) FINISHED
 => [internal] load build definition from Dockerfile                                                                                                                                                                                                                 0.0s
 => => transferring dockerfile: 32B                                                                                                                                                                                                                                  0.0s
 => [internal] load .dockerignore                                                                                                                                                                                                                                    0.0s
 => => transferring context: 2B                                                                                                                                                                                                                                      0.0s
 => [internal] load metadata for docker.io/library/fedora:31                                                                                                                                                                                                         1.4s
 => [internal] load metadata for docker.io/library/golang:1.15                                                                                                                                                                                                       1.3s
 => [build 1/7] FROM docker.io/library/golang:1.15@sha256:d141a8bca046ade2c96f89e864cd31f5d0ba88d5a71d62d59e0e1f2ecc2451f1                                                                                                                                           0.0s
 => CACHED [stage-1 1/2] FROM docker.io/library/fedora:31@sha256:ba4fe6a3da48addb248a16e8a63599cc5ff5250827e7232d2e3038279a0e467e                                                                                                                                    0.0s
 => [internal] load build context                                                                                                                                                                                                                                    0.5s
 => => transferring context: 43.06MB                                                                                                                                                                                                                                 0.5s
 => CACHED [build 2/7] WORKDIR /tmp/src                                                                                                                                                                                                                              0.0s
 => CACHED [build 3/7] COPY go.mod .                                                                                                                                                                                                                                 0.0s
 => CACHED [build 4/7] COPY go.sum .                                                                                                                                                                                                                                 0.0s
 => CACHED [build 5/7] RUN go mod download                                                                                                                                                                                                                           0.0s
 => [build 6/7] COPY . .                                                                                                                                                                                                                                             0.2s
 => [build 7/7] RUN ./hack/build.sh                                                                                                                                                                                                                                 12.5s
 => [stage-1 2/2] COPY --from=build /tmp/src/out/chart-verifier /app/chart-verifier                                                                                                                                                                                  0.1s
 => exporting to image                                                                                                                                                                                                                                               0.2s
 => => exporting layers                                                                                                                                                                                                                                              0.2s
 => => writing image sha256:7302e88a2805cb4be1b9e130d057bd167381e27f314cbe3c28fbc6cb7ee6f2a1                                                                                                                                                                         0.0s
 => => naming to quay.io/redhat-certification/chart-verifier:07e369d
```

The container image created by the build program is tagged with the commit ID of the working directory at the time of
the build: `quay.io/redhat-certification/chart-verifier:0d3706f`.

## Usage

### Local Usage

To verify a chart against all available checks:

```text
> out/chart-verifier verify ./chart.tgz
> out/chart-verifier verify ~/src/chart
> out/chart-verifier verify https://www.example.com/chart.tgz
```

To apply only the `is-helm-v3` check:

```text
> out/chart-verifier verify --enable is-helm-v3 https://www.example.com/chart.tgz
```

To apply all checks except `is-helm-v3`:

```text
> out/chart-verifier verify --disable is-helm-v3 https://www.example.com/chart.tgz
```

### Container Usage

The container image produced in 'Building chart-verifier' can then be executed with the Docker client
as `docker run -it --rm quay.io/redhat-certification/chart-verifier:0d3706f verify`.

If you haven't built a container image, you could still use the Docker client to execute the latest release available in
Quay:

```text
> docker run --rm quay.io/redhat-certification/chart-verifier:latest verify --help
Verifies a Helm chart by checking some of its characteristics

Usage:
  chart-verifier verify <chart-uri> [flags]

Flags:
  -S, --chart-set strings           set values for the chart (can specify multiple or separate values with commas: key1=val1,key2=val2)
  -G, --chart-set-file strings      set values from respective files specified via the command line (can specify multiple or separate values with commas: key1=path1,key2=path2)
  -X, --chart-set-string strings    set STRING values for the chart (can specify multiple or separate values with commas: key1=val1,key2=val2)
  -F, --chart-values strings        specify values in a YAML file or a URL (can specify multiple)
      --debug                       enable verbose output
  -x, --disable strings             all checks will be enabled except the informed ones
  -e, --enable strings              only the informed checks will be enabled
  -h, --help                        help for verify
      --kube-apiserver string       the address and the port for the Kubernetes API server
      --kube-as-group stringArray   group to impersonate for the operation, this flag can be repeated to specify multiple groups.
      --kube-as-user string         username to impersonate for the operation
      --kube-ca-file string         the certificate authority file for the Kubernetes API server connection
      --kube-context string         name of the kubeconfig context to use
      --kube-token string           bearer token used for authentication
      --kubeconfig string           path to the kubeconfig file
  -n, --namespace string            namespace scope for this request
  -V, --openshift-version string    set the value of certifiedOpenShiftVersions in the report
  -o, --output string               the output format: default, json or yaml
      --registry-config string      path to the registry config file (default "/home/baiju/.config/helm/registry.json")
      --repository-cache string     path to the file containing cached repository indexes (default "/home/baiju/.cache/helm/repository")
      --repository-config string    path to the file containing repository names and URLs (default "/home/baiju/.config/helm/repositories.yaml")
  -s, --set strings                 overrides a configuration, e.g: dummy.ok=false
  -f, --set-values strings          specify application and check configuration values in a YAML file or a URL (can specify multiple)

Global Flags:
      --config string   config file (default is $HOME/.chart-verifier.yaml)
```

To verify a chart on the host system, the directory containing the chart should be mounted in the container; for http or
https verifications, no mounting is required:

```text
> docker run --rm quay.io/redhat-certification/chart-verifier:latest verify https://github.com/redhat-certification/chart-verifier/blob/main/pkg/chartverifier/checks/chart-0.1.0-v3.valid.tgz?raw=true

```

Here is another example for a chart on the host system using volume mount. In
the below example, the chart is located in the current directory:

```text
> docker run -v $(pwd):/charts --rm quay.io/redhat-certification/chart-verifier:latest verify /charts/chart-0.1.0-v3.valid.tgz

```
