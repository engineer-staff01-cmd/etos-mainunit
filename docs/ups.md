# UPS

- `apcupsd` と呼ばれるデーモンで APC 社製の無停電電源装置を監視している
- `systemctl status apcupsd.service` でデーモンの状態を確認できる
- `apcaccess` コマンドで UPS のステータスを確認可能

## UPS の設定

- `/etc/apcupsd/apcupsd.conf` にデーモンの設定が記載されている
  - `DEVICE` の項目のみデフォルトから修正している
- 電源遮断時に自動でシャットダウンする設定、 `TIMEOUT` が存在するが、シャットダウンを無効にするため `0` に設定する