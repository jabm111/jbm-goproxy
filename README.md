### Building

- Build for development: `./bin/dev`
- Build for production: `./bin/build_linux`

### Evironment Variables

You can customize the behaviour of the reverse proxy using the environment variables below.

Environment Variable | Required | Default Value | Description
--- | --- | --- | ---
GO_PROXY_SCHEME | no | http | The scheme of the server being proxied (http/https).
GO_PROXY_HOST | no | localhost:8080 | The host of the server being proxied.
GO_STATIC_DIR | no | static | The directory containing the static assets to be served directly (not proxied).
GO_STATIC_PREFIX | no | static | The path prefix for static assets.
GO_PORT | no | :8888 | The port to run the go server on. If :443 is used TLS is enabled and GO_DOMAINS is required.
GO_DOMAINS | if GO_PORT == :443 |  | The domain(s) (1 or 2) comma separated that the go server will automatically provision certificates for (ex: yourdomain.com,www.yourdomain.com).
GO_CERT_CACHE_DIR | no | /home/gouser/letsencrypt | The location where the certificates will be stored on the server.
