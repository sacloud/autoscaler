# E2Eテスト追加時の注意点

.gitignoreで`autoscaler.yaml`を無視しているため、明示的に追加が必要です。

    git add -f e2e/xxx/autoscaler.yaml
    git commit ...