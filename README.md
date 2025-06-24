# External Proxy

Simple external proxy server for k8s. It implements reverse proxy and tls termination on edge server outside k8s cluster but using k8s annotations for configuration.

## Install

```sh
curl -L https://github.com/egor-lukin/external-proxy/releases/download/v0.0.1/external-proxy -o external-proxy
sudo mv external-proxy /usr/local/bin/external-proxy 
sudo chmod +x /usr/local/bin/external-proxy

```

## Usage

- Run proxy

``` sh
external-proxy run --certsPath=certsPath --kubeNamespace=default --interval=10s --nginxSettingsPath=/etc/nginx/sites-enabled
```

## Release

To create a new release and trigger the GitHub Actions workflow to build and upload the binary, create and push a git tag with the desired version. For example:

```sh
git tag v0.0.1
git push origin v0.0.1
```
