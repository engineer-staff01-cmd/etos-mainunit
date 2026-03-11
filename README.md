# etos-mainunit

イートス主装置の親機プログラム

## 接続先環境

ビルド方法によって、クラウドの接続先を切り替えています。

ステージング環境に接続

* `go run` で実行
* `sh build.sh` でビルド

本番環境に接続

* `sh build.sh release` でビルド

## Build

以下のコマンドを実行することで Armadillo 用のバイナリを生成できます。

```bash
# debug build(Staging, CSV)
sh build.sh

# release build(Staging, Sensor)
sh build.sh sens_debug

# release build(Product, CSV)
sh build.sh test

# release build(Product, Sensor)
sh build.sh release
```

## Test

以下のコマンドを実行することでテストを実行できます

```bash
sh test.sh
```

## Archive file

生成された zip ファイルは以下のような構成になっています。

```
.
├ etos-mainunit 実行バイナリ
└ md5sum        チェックサム(md5)
```

あらかじめ上記手順でビルドを行い、以下のコマンドでアーカイブを生成できます。

```bash
sh archive.sh
```

生成したアーカイブの中身を確認したい場合、以下のコマンドでアーカイブを展開できます。

```bash
sh unarchive.sh
```

## APIの定義更新

`api` フォルダに Ghost社が開発しているAPIの定義ファイル (OpenAPI) を配置してある。

- api/
    - syncapi.json
    - measureapi.json

上記のフォルダに最新の定義ファイルを配置後、以下のコマンドを実行することで、 `syncapi`, `measureapi` フォルダ内の定義ファイルを更新することができる。

```sh
sh update-api.sh
```

## Documents

- [アラート](docs/alert.md)
- [ログ](docs/log.md)
- [スリープモード(廃止)](docs/sleep-mode.md)
- [UPS](docs/ups.md)
