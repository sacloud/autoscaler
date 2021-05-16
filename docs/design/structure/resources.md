## Resources

CoreのConfigurationの中で指定されるリソース

### サーバ

```yaml
- Server: 
  - DedicatedCPU: boolean
  - PrivateHostID: types.ID
  - Zone: "is1a"
  - Plans: # 省略した場合、AutoScaler側で定義されたプラン一覧から順次選択される(例: C1M1 -> C2M4 -> C4M8) 
      - core: 1
        memory: 1
      - core: 2
        memory: 4
      - core: 4
        memory: 8
```

### オートスケーリンググループ(サーバ)
```yaml
- ServerGroup:
  - Selector: types.Tags
  - MaxSize: number # 最大インスタンス数
  - MinSize: number # 最小インスタンス数
  - Zones: # 対象ゾーン、ここで指定した順にリソースが分散される
    - is1a
    - is1b
    - tk1a
    - tk1b
  # サーバのテンプレート
  - ServerTemplate:
    - OSType: libsacloud/sacloud/ostype.ArchiveOSType
    - ArchiveID: types.ID   # ostype=Customの場合に使う
    - StartupScript: string # 文字列でスタートアップスクリプトを指定する
    - Core: number
    - Memory: number
    - DedicatedCPU: boolean
    - PrivateHostID: types.ID
    - Plans:
        - core: 1
          memory: 1
  # サーバの外側のリソース群(ネスト可能)
  - Wrappers 
    - Type: EnhancedLoadBalancer
      Selector: types.Tags
      Wrappers: ...

```

### エンハンスドロードバランサ
```yaml
# エンハンスドロードバランサ
- EnhancedLoadBalancer:
    Plans: # 省略した場合、AutoScaler側で定義されたプラン一覧から順次選択される(例: 100 -> 1000 -> 10000) 
      - 100
      - 1000
      - 10000
      - 100000  
```