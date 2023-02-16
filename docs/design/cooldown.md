# cooldownの基準にクラウド上のリソースのmodified_atを用いる

- URL: https://github.com/sacloud/autoscaler/issues/466
- Author: @yamamoto-febc

## 概要

autoscalerではup/down操作を意図せず連続実行してしまうことを防ぐために冷却期間(cooldown)というパラメータを設けている。
```yaml
# cooldownの設定例
autoscaler:
  cooldown: 300 # 前回up/downが実行されてから300秒間はup/downの実行を抑止する
```

cooldown期間中であるかの判定は前回up/downを実行した時間と現在時刻を比較することで行われている。  
この時の前回up/downを実行した時間はインメモリに保持されている。

このためメンテナンスやバージョンアップなどでautoscalerを再起動した場合、保持していた前回up/down実行時刻が消えてしまうため、意図しないタイミングでup/down操作が行えるようになってしまう。

例: `cooldown:300`とした状態でスケールアップを実施、すぐにautoscalerを再起動すると300秒待たずともup/downが行えるようになる。

この問題を解決するため、さくらのクラウドAPIを呼び出してスケール対象リソース(群)の情報を取得し、modified_atフィールドを参照することで判定するように変更する。

## 実装方針

#### 実装すること

- cooldown判定時にさくらのクラウドAPIを呼び出してリソースを取得〜modified_atフィールドを参照しての判定

#### 実装しないこと

- キャッシュなどを用いた負荷軽減周り(主にさくらのクラウドAPI呼び出し周りの負荷軽減)

Note: 毎回のup/down実行時にさくらのクラウドAPI呼び出しが発生するが、GETリクエストのみであることや、そこまで高頻度にup/downを呼び出すユースケースが考えづらいことからAPI呼び出し負荷増大を許容する。

## 実装

- `core.ResourceDefinition`インターフェースにfunc`LastModifiedAt(...)`を追加し、各定義ごとに実リソース(群)からのmodified_atを取得可能にする
- `core.ResourceDefinitions`にも`LastModifiedAt(...)`を追加し、内包するResourceDefinitionsが返したmodified_atの内、最も遅い時間を示すmodified_atの値を返すようにする
- `core.JobStatus`のfunc`Acceptable(...)`を上記のLastModifiedAt()を呼び出して判定するよう修正


### 更新内容

- 2023/2/16: 作成