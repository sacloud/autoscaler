# sacloud/autoscalerのデータ構造

## Inputs

Inputsは外部からのリクエストなどを契機にCoreを呼び出す。

- [InputsからCoreへのリクエスト定義](../../request.proto)
- [Grafana Inputs](inputs_grafana.md)
- [AlertManager Inputs](inputs_alertmanager.md)

## Core

Coreはコマンドラインオプションや設定ファイルなどから起動構成を読み取り起動、Inputsからのリクエストを受けてHandlersを呼び出す。  

- [Coreのコンフィギュレーション](core.md)
- [リソース定義](resources.md)
- [CoreからHandlersへのリクエスト定義](../../handler.proto)

