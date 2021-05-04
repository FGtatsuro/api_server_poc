# API Server PoC

## 検証環境

- OSX(10.15.7)
- Docker(Engine: 20.10.5)
- GNU Make(3.81)

## タスク

### APIサーバのビルド

```bash
$ make build
```

### APIサーバの起動

```bash
$ make start
```

### APIサーバの停止

```bash
$ make stop
```

### specテストの実行

```bash
$ make spec
```

## APIサーバの仕様

- 対応プロトコル: `HTTP`(`HTTPS`には対応していない)

(*)リクエストサンプルの実行には、以下のツールを使用する。

- [Curl](https://github.com/curl/curl)
- [jq](https://github.com/stedolan/jq)

### `/api`

- GET: `http://httpbin.org/headers` へのGETアクセス結果を返す

```json
$ curl -I -X GET '127.0.0.1:8080/api' -H 'authorization: tmp'
HTTP/1.1 200 OK
Content-Length: 219
Content-Type: text/plain; charset=utf-8
...
$ curl -s -X GET '127.0.0.1:8080/api' -H 'authorization: tmp' | jq '.'
{
  "headers": {
    "Accept-Encoding": "gzip",
    "Authorization": "tmp",
    "Host": "httpbin.org",
    "User-Agent": "Go-http-client/1.1",
    "X-Amzn-Trace-Id": "Root=1-60911f5e-30da301d0e95779a0cd4fa2a"
  }
}
```

- `authorization` ヘッダがない場合、Status 401を返す

```
$ curl -I -X GET '127.0.0.1:8080/api'
HTTP/1.1 401 Unauthorized
Content-Type: text/plain; charset=utf-8
Content-Length: 25
...
```
