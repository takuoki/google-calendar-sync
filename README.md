# google-calendar-sync

google-calendar-sync は、Google カレンダーのイベントとデータベースを同期するためのサンプルアプリケーションです。

## Requirement

- Google Cloud のサービスアカウントが用意されていること。
- 上記のサービスアカウントがアクセス可能な Google カレンダーが用意されていること。

## Get Started

1. リポジトリをクローンします。

```sh
git clone https://github.com/takuoki/google-calendar-sync.git
cd google-calendar-sync
```

2. サービスアカウントの鍵ファイル（JSON 形式）を、下記に配置します。

- api/credentials.json

3. アプリケーションを起動します。

```sh
make run
```

3. Google カレンダーを登録します。

```sh
curl --location --request POST 'http://localhost:8080/api/calendars/sample@sample.com/?name=sample'
```

4. Google カレンダーを同期します。

```sh
curl --location --request POST 'http://localhost:8080/api/sync/sample@sample.com/' \
  --header 'X-Goog-Resource-State: exists'
```

- 初回は存在する全データの同期、2 回目以降は差分データのみの同期となります。

5. データベースに接続し、結果を確認します。

```sh
mysql -h 127.0.0.1 -P 3306 -u appuser -ppassword -D app
```

## Additional Information

- Google カレンダーの予定の更新を検知し、Webhook を発火させる仕組みがありますが、ローカル環境では確認できません。
