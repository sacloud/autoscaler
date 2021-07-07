# Direct Inputs

Direct Inputsを利用する方法ついて記載します。  

## Direct Inputsとは

コマンドラインからAutoScaler Coreへ直接Up/Downリクエストを送信するためのCLIです。  
cronでの実行や動作確認などで利用します。

## 前提条件

- Direct InputsからAutoScaler Coreへの疎通が可能なこと

## 利用方法

`autoscaler inputs direct`で実行可能です。  
指定可能なオプションなどは以下の通りです。  

```shell
$ autoscaler inputs direct -h

Send Up/Down request directly to Core server

Usage:
  autoscaler inputs direct {up | down} [flags]...

Flags:
      --desired-state-name string    Name of the desired state defined in Core's configuration file
      --dest string                  Address of the gRPC endpoint of AutoScaler Core (default "unix:autoscaler.sock")
  -h, --help                         help for direct
      --resource-name string         Name of the target resource (default "default")
      --source string                A string representing the request source, passed to AutoScaler Core (default "default")
      --config string                Filepath to Inputs additional configuration file
      
Global Flags:
      --log-format string   Format of logging to be output. options: [ logfmt | json ] (default "logfmt")
      --log-level string    Level of logging to be output. options: [ error | warn | info | debug ] (default "info")
```

## TLS関連設定

[Inputs共通設定](../config.md)を参照ください。