# AutoScaler

![logo.svg](./docs/images/logo.svg)
[sacloud/autoscaler](https://github.com/sacloud/autoscaler) はさくらのクラウド上のリソースのオートスケーリングを行うためのツールです。

## Overview

![architecture.png](./docs/images/architecture.png)

sacloud/autoscalerはGrafanaやAlertManagerなどの監視ツールからのWebhookを受け、あらかじめ定義しておいたコンフィギュレーションに沿ってさくらのクラウド上のリソースのオートスケールを行います。  
オートスケールに際し、サーバの上流にロードバランサが存在する場合はロードバランサからのデタッチ/アタッチも行います。

以下のオートスケーリングに対応しています。

- `elb-vertical-scaler`: エンハンスドロードバランサの垂直スケール(CPSの変更)
- `elb-servers-handler`: エンハンスドロードバランサ配下のサーバのデタッチ/アタッチ
- `gslb-servers-handler`: GSLB配下のサーバのデタッチ/アタッチ
- `router-vertical-scaler`: ルータの垂直スケール(帯域幅の変更)
- `server-vertical-scaler`: サーバの垂直スケール(CPU/メモリサイズの変更)
- (近日実装予定)`server-horizontal-scaler`: サーバの水平スケール

## Getting Started

[Getting Started Guide](./docs/getting_started.md)を参照してください。

## License

`sacloud/autoscaler` Copyright (C) 2021 The sacloud Authors.

This project is published under [Apache 2.0 License](LICENSE.txt).
