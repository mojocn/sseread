# FAAS HTTP API

本文档描述项目部署到 Vercel 或 Netlify 后提供的 HTTP API。

## 基础 URL

将下文中的 `<base-url>` 替换为实际部署地址：

| 平台    | 基础 URL 示例                  |
| ------- | ------------------------------ |
| Vercel  | `https://<project>.vercel.app` |
| Netlify | `https://<site>.netlify.app`   |

Netlify 同时提供两种访问路径：

- 推荐使用项目根路径：`/netdog-http`、`/netdog-network`、`/ip`
- 函数原始路径：`/.netlify/functions/netdog-http`、`/.netlify/functions/netdog-network`、`/.netlify/functions/ip`

## 接口总览

| 方法   | Vercel 路径       | Netlify 路径      | 说明                                |
| ------ | ----------------- | ----------------- | ----------------------------------- |
| `POST` | `/netdog-http`    | `/netdog-http`    | 检查 HTTP/HTTPS 目标                |
| `POST` | `/netdog-network` | `/netdog-network` | 检查 TCP/UDP 目标                   |
| `GET`  | `/ip`             | `/ip`             | 获取客户端请求头和 Netlify 地理信息 |

所有探测接口均返回 `Content-Type: application/json; charset=utf-8`。当前实现会以 HTTP `200 OK` 返回探测结果，即使探测结果中的 `error` 不为空。

## 1. HTTP 探测

### 请求

```http
POST <base-url>/netdog-http
Content-Type: application/json
```

请求体：

| 字段      | 类型    | 必填 | 说明                                                                                 |
| --------- | ------- | ---- | ------------------------------------------------------------------------------------ |
| `method`  | string  | 否   | 目标 HTTP 方法，例如 `GET`、`POST`。省略时由 Go HTTP 客户端使用默认方法。            |
| `url`     | string  | 是   | 目标 URL。                                                                           |
| `headers` | object  | 否   | 目标请求头，键和值均为字符串。                                                       |
| `body`    | string  | 否   | 目标请求体的 Base64 编码。因为服务端字段类型为 `[]byte`，JSON 不接受普通明文字符串。 |
| `timeout` | integer | 否   | 超时时间，单位为纳秒。小于 1 秒按 1 秒处理，大于 60 秒按 60 秒处理。                 |

示例：向目标发送 JSON 请求。`eyJwaW5nIjoid29ybGQifQ==` 是 `{"ping":"world"}` 的 Base64 编码。

```bash
curl -X POST "<base-url>/netdog-http" \
	-H "Content-Type: application/json" \
	-d '{
		"method": "POST",
		"url": "https://example.com/health",
		"headers": {
			"Accept": "application/json",
			"Content-Type": "application/json"
		},
		"body": "eyJwaW5nIjoid29ybGQifQ==",
		"timeout": 5000000000
	}'
```

### 响应

```json
{
  "cost": 183456789,
  "tls_issuer": "CN=Example CA",
  "tls_subject": "CN=example.com",
  "tls_not_before": "2026-01-01T00:00:00Z",
  "tls_not_after": "2027-01-01T00:00:00Z",
  "error": null,
  "headers": {
    "Content-Type": ["application/json"],
    "Content-Length": ["16"]
  },
  "body": "eyJzdGF0dXMiOiJvayJ9"
}
```

响应字段：

| 字段             | 类型        | 说明                                                  |
| ---------------- | ----------- | ----------------------------------------------------- |
| `cost`           | integer     | 探测耗时，单位为纳秒。                                |
| `tls_issuer`     | string      | TLS 证书签发者；未进行 TLS 握手时为空字符串。         |
| `tls_subject`    | string      | TLS 证书主题；未进行 TLS 握手时为空字符串。           |
| `tls_not_before` | string      | TLS 证书生效时间，RFC 3339 格式；无证书时为零值时间。 |
| `tls_not_after`  | string      | TLS 证书过期时间，RFC 3339 格式；无证书时为零值时间。 |
| `error`          | string/null | 探测错误信息；成功时为 `null`。                       |
| `headers`        | object      | 目标响应头。                                          |
| `body`           | string      | 目标响应体的 Base64 编码，最多读取 2 MiB。            |

## 2. TCP/UDP 网络探测

### 请求

```http
POST <base-url>/netdog-network
Content-Type: application/json
```

```json
{
  "network": "tcp",
  "host": "example.com",
  "port": "443",
  "timeout": 5000000000,
  "tls": true
}
```

请求字段：

| 字段      | 类型    | 必填 | 说明                                                                       |
| --------- | ------- | ---- | -------------------------------------------------------------------------- |
| `network` | string  | 是   | `tcp` 或 `udp`，大小写不敏感。                                             |
| `host`    | string  | 是   | 主机名或 IP 地址。                                                         |
| `port`    | string  | 是   | 目标端口，例如 `80`、`443`。                                               |
| `timeout` | integer | 否   | 超时时间，单位为纳秒，范围会限制为 1 到 60 秒。                            |
| `tls`     | boolean | 否   | 为 `true` 时，在 TCP 连接后执行 TLS 握手并提取证书信息。UDP 会忽略该字段。 |

### 响应

```json
{
  "cost": 42123456,
  "tls_issuer": "CN=Example CA",
  "tls_subject": "CN=example.com",
  "tls_not_before": "2026-01-01T00:00:00Z",
  "tls_not_after": "2027-01-01T00:00:00Z",
  "error": null
}
```

字段含义与 HTTP 探测响应中的同名字段相同。`tls` 为 `false` 或使用 UDP 时，TLS 字段保持零值。

## 3. 客户端 IP 和请求信息

### 请求

```http
GET <base-url>/ip
```

### Vercel 响应

Vercel 版本返回纯文本客户端地址，优先使用 `X-Forwarded-For`，否则使用 `RemoteAddr`：

```http
200 OK
Content-Type: text/plain; charset=utf-8

203.0.113.10
```

### Netlify 响应

Netlify 版本返回 JSON，并包含平台注入的请求头信息：

```json
{
  "x-forwarded-for": "203.0.113.10",
  "x-nf-client-connection-ip": "203.0.113.10",
  "x-country": "US",
  "x-language": "en",
  "user-agent": "Mozilla/5.0",
  "sec-ch-ua-platform": "macOS",
  "sec-ch-ua-mobile": "?0",
  "geo": {
    "city": "San Francisco",
    "country": "US"
  }
}
```

字段缺失时，对应值可能为空字符串；`geo` 只有在 Netlify 的 `x-nf-geo` 头存在且能正确解码时才会返回。

## 错误和限制

- 请求体不是合法 JSON 时，`/netdog-http` 和 `/netdog-network` 仍返回 `200 OK`，响应中包含 `error` 字段。
- 目标连接失败、URL 无效、端口不可达或 TLS 握手失败时，错误写入 `error`，不会改变外层 HTTP 状态码。
- HTTP 目标响应体最多读取 2 MiB，超出部分会被截断。
- HTTP 探测会覆盖目标请求的 `User-Agent`，使用项目内置的移动端 Safari 标识。
- 探测功能可访问部署环境能够访问的任意目标地址；生产环境建议在调用方增加认证、限流和目标地址白名单。
