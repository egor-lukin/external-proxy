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

- Configuration

For setup simple https reverse proxy for custom route implement next steps:

1. Create k8s ingress with specific annotations

``` yaml
external-proxy/domain: example.com,
external-proxy/server-snippets: |
 server {
    listen 443 ssl;
    server_name example.com;

    ssl_protocols TLSv1 TLSv1.1 TLSv1.2;
    ssl_certificate {{ .CertsPath }}/example.com.crt;
    ssl_certificate_key {{ .CertsPath }}/example.com.key;

    location / {
        proxy_http_version 1.1;
        proxy_ssl_server_name on;
        proxy_ssl_verify off;
        proxy_read_timeout     60;
        proxy_connect_timeout  60;
        keepalive_requests 8192;

        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        proxy_pass https://example.com:443;
    }
  }

```

2. Create k8s tls secret with specific annotaitions:

``` yaml
external-proxy/domain: example.com
```

## Release

To create a new release and trigger the GitHub Actions workflow to build and upload the binary, create and push a git tag with the desired version. For example:

```sh
git tag v0.0.1
git push origin v0.0.1
```
