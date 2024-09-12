Automatically copies new/changed files to kubernetes deployment

## Building and installing

```bash
go build ./... -o auto-k8s-cp
sudo install auto-k8s-cp /usr/local/bin
```

## Usage

```
auto-k8s-cp FLAGS

FLAGS:
Usage of /tmp/go-build1860283438/b001/exe/main:
  --container string
        Container name in the deployment's pods
  --container-path string
        Path inside the container to copy files to
  --deployment string
        Kubernetes deployment name
  --local-path string
        Local path to monitor for file changes
  --namespace string
        Kubernetes namespace
```

Example:

```bash
auto-k8s-copy --local-path="." --deployment="log-test9-patchdemo" --container="patchdemo" --container-path="/var/www/html" --namespace="default"
```

## License

MDGPL