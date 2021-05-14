# sacloud/autoscalerのデータ構造

## Inputs

Inputsは外部からのリクエストなどを契機にCoreを呼び出す。

- [InputsからCoreへのリクエスト定義](inputs-to-core)
- [Grafana Inputs](inputs_grafana)
- [AlertManager Inputs](inputs_alertmanager)

## Core

Coreはコマンドラインオプションや設定ファイルなどから起動構成を読み取り起動、Inputsからのリクエストを受けてHandlersを呼び出す。  

- [Coreのコンフィギュレーション](core)
- [リソース定義](resources)
- [CoreからHandlersへのリクエスト定義](core-to-handlers)

