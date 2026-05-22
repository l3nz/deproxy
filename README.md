# deproxy

An intercepting HTTP/HTTPS proxy for inspection and request replacement.

## Build

```sh
GOROOT= go build .
```

## Usage

```sh
./deproxy              # listen on :8080
./deproxy -addr :9090  # custom port
./deproxy -v           # verbose (logs goproxy internals)
```

## Firefox setup

HTTPS interception requires importing the proxy's CA certificate into Firefox.

**1. Export the CA cert:**

```sh
./deproxy -dump-cert goproxy-ca.pem
```

**2. Import into Firefox:**

- Open Firefox → Settings → search "certificates" → **View Certificates**
- Go to the **Authorities** tab → click **Import** → select `goproxy-ca.pem`
- Check **Trust this CA to identify websites** → OK

**3. Configure Firefox to use the proxy:**

- Settings → **Network Settings** → **Manual proxy configuration**
- HTTP Proxy: `127.0.0.1`, Port: `8080`
- Check **Also use this proxy for HTTPS**

**4. Run deproxy:**

```sh
./deproxy
```

Traffic will appear in the terminal as:

```
-> GET https://example.com/path
<- 200 https://example.com/path
```

## Intercepting requests

Edit `main.go` and add `OnRequest` handlers before the `HandleConnect` line.
The `matchHost` helper matches any URL whose host contains the given substring:

```go
proxy.OnRequest(matchHost("example.com")).DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
    resp := goproxy.NewResponse(req, "text/plain", http.StatusOK, "intercepted")
    return req, resp
})
```
