# ログ

- このプログラムのログはすべて `syslog` に送信される。
  - タグは `ecoRAMDER`
- (Priority = facility(LOG_USER) + severity(LOG_NOTICE)) として出力
- 実装には Go の標準パッケージ [`log/syslog`](https://pkg.go.dev/log/syslog) を使用している。
- ログのメッセージには Prefix としてエラーレベルを付与している
  - `grep <ERROR_LEVEL> <log-file>` などでエラーレベルごとに絞り込むことができる
- エラーレベル一覧
  - `EMERG` 未使用
  - `ALERT` 未使用
  - `CRITICAL` 仕様上想定されていない致命的な問題が発生した時に出力される。
  - `ERROR` 仕様上想定されているエラー
  - `WARNING` 動作上問題ないが、警告するために使用される。
  - `NOTICE` 未使用
  - `INFO` 異常は発生していないが、情報を出力するために使用される
  - `DEBUG` デバッグ時に挙動をトレースするために出力される。

## デバッグ

### Windows の WSL 上などでデバッグする場合

`rsyslog` デーモンが起動していない場合、panic が発生するためあらかじめ以下のコマンドで起動しておく

```
systemctl start rsyslog
```

panic 例

```
panic: Unix syslog delivery error

goroutine 1 [running]:
etos-mainunit/command.NewUnitLog({0x7cb171, 0x9})
        /mnt/c/Users/ASTINA/go/src/github.com/etos-mainunit/command/syslog.go:36 +0xf6
etos-mainunit/command.init()
        /mnt/c/Users/ASTINA/go/src/github.com/etos-mainunit/command/syslog.go:26 +0x3ff
FAIL    etos-mainunit/command   0.010s
FAIL
```

## ログの保存先

- SDカードによるブートが有効になっている前提
  - デフォルトの `/var/log/` 配下にログが保存される
    - `tail -f /log/syslog/user.log` で確認できる

## ログのローテーション

ログのローテーションは `logrotated` によって行われる。  
設定は以下のようになっている

- `maxsize 200M` 1 ファイルの最大サイズは 200M
- `rotate 14` 現在のログ 1 つ + バックアップで 14 つ。合計 15 ファイルを保存しておく
- `daily` 1 日ごとにローテートを行う。(ログが `maxsize` を超えた場合、ローテートされる)
- `delaycompress` 一番新しいバックアップファイルは gz 圧縮を行わない
- `compress` バックアップファイルを圧縮する

[設定ファイル](../etc/logrotate.d/rsyslog)

```
/var/log/mail.info
/var/log/mail.warn
/var/log/mail.err
/var/log/mail.log
/var/log/daemon.log
/var/log/kern.log
/var/log/auth.log
/var/log/user.log
/var/log/lpr.log
/var/log/cron.log
/var/log/debug
/var/log/messages
{
        maxsize 200M
        rotate 14
        daily
        missingok
        notifempty
        compress
        delaycompress
        sharedscripts
        postrotate
                /usr/lib/rsyslog/rsyslog-rotate
        endscript
}
```
