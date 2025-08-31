# Google Calendar API CLI

このCLIツールは、Google Calendar API の主要な機能の挙動確認や操作を行うための汎用的な CLI ツールです。

## 機能一覧

- **EventsInstancesCall**: 繰り返しイベントのインスタンス一覧を取得
- **EventsListCall**: カレンダーのイベント一覧を取得
- **ChannelsStopCall**: Push通知チャンネルの停止
- **EventsWatchCall**: イベントの変更通知（Push通知）の購読開始


## インストール

以下のコマンドでCLIをインストールできます。

```sh
go install github.com/takuoki/google-calendar-sync/cli/gcalcli@latest
```

## 使い方

### 認証情報の準備

デフォルトでは `cli/credentials.json` で指定されたサービスアカウントを利用します。
このサービスアカウントがカレンダーを参照できるよう、事前に権限を付与してください。

#### OAuth 2.0 Support

`--oauth` オプションを付与すると、以下の環境変数を利用して認証します。この場合、`--refresh-token` オプションも必須です。

- `OAUTH_CLIENT_ID`
- `OAUTH_CLIENT_SECRET`
- `OAUTH_REDIRECT_URL`
- `--refresh-token`（コマンドライン引数）

例:

```sh
export OAUTH_CLIENT_ID=xxxx.apps.googleusercontent.com
export OAUTH_CLIENT_SECRET=xxxx
export OAUTH_REDIRECT_URL=https://yourapp.com/oauth2callback

./gcalcli --oauth --refresh-token <リフレッシュトークン> ...
```

### 実行例

#### イベント取得
```sh
./gcalcli event --calendar-id <カレンダーID> --event-id <イベントID>
```

#### イベント一覧取得
```sh
./gcalcli events --calendar-id <カレンダーID>
```

#### 定期イベントのインスタンス取得
```sh
./gcalcli instances --calendar-id <カレンダーID> --event-id <イベントID>
```

#### イベント変更通知のサブスクライブ開始
```sh
./gcalcli watch --calendar-id <カレンダーID> --channel-id <チャンネルID> --address <通知先URL>
```

#### イベント変更通知のサブスクライブ終了
```sh
./gcalcli stop --channel-id <チャンネルID> --resource-id <リソースID>
```
