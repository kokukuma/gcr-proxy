# 実験
## 背景
+ GCRを使いたいが, GCPリテラシーがそれほど高くない状況では, GCPアカウントを作ってもらうとか, gcloudコマンド入れる時点で抵抗ある. 特にwindows.
+ GCRのreverse proxyを作って, 認証を肩代わりさせてみる.
+ docker deamonからは, GCRを直接見ずに, このproxyをRegistryとしてみる.
+ もちろん, GCPアカウント作って, gcloudコマンド入れる方が正しい.

## やりかた
+ 基本的には, docker deamonとGCRのやり取りをそのまま転送する.
+ ただし認証の部分だけ書き換える
  + `/v2/` のResponse: `Www-Authenticate` gcr.ioをproxyのホスト名に変える.
  + `/v2/token` のRequest: `Authorization` GCR使えるサービスアカウントのjsonkeyに変える.

## 使い方
+ docker clientでは適当にdocker loginしておく.
  + 認証ありの状態にしないと, `/v2/token`叩いてくれないので.
  ```
  docker login -u gcr-proxy -p password gcrproxy.domain
  ```
+ 後はgcr.ioをproxyのホスト名に変えて, 使う.
  ```
  docker pull gcr.io/test/kokukuma
  ↓
  docker pull gcrproxy.domain/test/kokukuma
  ```

## 立ち上げ方
### https
+ いろいろホストに置く
  + サーバ証明書, 秘密鍵, サービスアカウントのJson Key.
  + 下の例では, /Users/karino-t/sa/以下に置いてある.
  + k8sでは, secretに登録してmountする. k8s上で使わないと思うけど.
+ build
  ```
  docker build -t gcrproxy .
  ```
+ run
  ```
  docker run -e CRT_PATH="/sa/CERT.pem" \
             -e KEY_PATH="/sa/PRIVATE-KEY-Dec.pem" \
             -e SERVICE_ACCOUNT_PATH="/sa/gcr-proxy-sa.json" \
             -e PROXY_AUTH=${PROXY_AUTH} \
             -e PROXY_URL=${PROXY_URL_LOCAL} \
             -e REGISTRY_URL=${REGISTRY_URL} \
             -v /Users/karino-t/go/src/github.com/kokukuma/gcr-proxy/cert:/sa/ \
             -p 443:8000 \
             gcrproxy
  ```

### k8s
```
kubectl apply -f k8s/deployment.yml
kubectl apply -f k8s/service.yml
envsubst < k8s/secret.yml | kubectl apply -f -
```
