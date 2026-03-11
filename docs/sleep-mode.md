# スリープモード(廃止)

Armadillo は スリープ機能をサポートしているが、このシステムでは利用していない。  
その代わりにプログラム内で疑似スリープ機能を実装している。

## 疑似スリープ機能に関して

- 電源がバッテリー駆動に切り替わった際にスリープ機能が有効になる
- スリープ中はデータベースのアクセスや機器の制御、アラート送信などは行われない
- スリープ中は `apcaccess` コマンドの実行結果に含まれる `STATUS` を 10 秒間隔でチェックしている。
  - `ONLINE` UPS が電源に接続されている
  - `ONBATT` UPS がバッテリーから電源を供給している

## Armadillo のスリープ機能に関して調査メモ

### Armadillo のスリープ機能を使用しなかった理由

- Linux カーネルがサスペンド状態になるため、
- ウェイクアップ要因による割込みでは電源復旧時に自動でスリープモードを解除する要求を満たすことができない

### スリープ機能に移行する方法

`standby` または `mem` を `/sys/power/state` に書き込むことでスリープモードに移行できる。

```
echo -ne "standby" > /sys/power/state
```

### USB による抜き差しをウェイクアップ要因に設定する方法

`standby` または `mem` 両方で復帰可能

```
echo enabled > /sys/bus/platform/devices/30b10000.usb/power/wakeup
echo enabled > /sys/bus/platform/drivers/ci_hdrc/ci_hdrc.0/power/wakeup
echo enabled > /sys/bus/platform/drivers/ci_hdrc/ci_hdrc.0/usb1/power/wakeup
```

### 参考リンク

- UPS
  - [BR400_500G-JP_Manual.pdf | Powered by Box](https://schneider-electric.app.box.com/s/ogy41o5yexywc5a5ia3dztxrd8tcl7dw)
- Armadillo
  - [第 7 章 Linux カーネル仕様](https://manual.atmark-techno.com/armadillo-640/armadillo-640_product_manual_ja-1.11.1/ch07.html#sct.kernel_spec-pm)
  - [第 9 章 Linux カーネルデバイスドライバー仕様](https://manual.atmark-techno.com/armadillo-4x0/armadillo-400_series_software_manual_ja-1.9.2/ch09.html#sec-power-man)
