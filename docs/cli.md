# AutoScaler CLIリファレンス

## 利用方法/ヘルプ

`autoscaler -h`でヘルプが表示されます。

```shell
$ autoscaler -h

autoscaler is a tool for managing the scale of resources on SAKURA cloud

Usage:
  autoscaler [command]

Available Commands:
  completion  Generate completion script
  help        Help about any command
  inputs      A set of sub commands to manage autoscaler's inputs
  server      A set of sub commands to manage autoscaler's core server
  version     show version

Flags:
  -h, --help                help for autoscaler
      --log-format string   Format of logging to be output. options: [ logfmt | json ] (default "logfmt")
      --log-level string    Level of logging to be output. options: [ error | warn | info | debug ] (default "info")

Use "autoscaler [command] --help" for more information about a command.
```

指定可能なサブコマンドやオプションは各コマンドのヘルプを参照してください。

## シェル補完

sacloud/autoscalerはシェル補完に対応しています。  
シェル補完の有効化方法はご利用のシェルごとに異なります。

`autoscaler completion --help`で表示される手順に従ってください。  
