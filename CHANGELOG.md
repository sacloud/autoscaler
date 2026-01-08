# Changelog

## [v0.19.3](https://github.com/sacloud/autoscaler/compare/v0.19.2...v0.19.3) - 2026-01-08
- use default http.Client by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/687
- e2e: use default http.Client by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/689

## [v0.19.2](https://github.com/sacloud/autoscaler/compare/v0.19.1...v0.19.2) - 2026-01-08
- iaas-service-go v1.21.1 & go 1.25.5 by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/685

## [v0.19.1](https://github.com/sacloud/autoscaler/compare/v0.19.0...v0.19.1) - 2025-12-11
- update dependencies - iaas-service-go v1.20.1 and iaas-api-go v1.23.1 by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/682

## [v0.19.1](https://github.com/sacloud/autoscaler/compare/v0.19.0...v0.19.1) - 2025-12-11
- update dependencies - iaas-service-go v1.20.1 and iaas-api-go v1.23.1 by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/682

## [v0.19.0](https://github.com/sacloud/autoscaler/compare/v0.18.2...v0.19.0) - 2025-12-08
- ci: bump actions/setup-go from 5 to 6 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/658
- go: bump github.com/spf13/cobra from 1.9.1 to 1.10.1 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/654
- iaas-service-go v1.18.1 by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/673
- use terraform-provider-sakuracloud v2.31.2 by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/674
- Fix incorrect request type in Down() logger by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/677
- ci: bump actions/checkout from 5 to 6 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/676
- go: bump google.golang.org/grpc from 1.75.0 to 1.76.0 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/675

## [v0.18.2](https://github.com/sacloud/autoscaler/compare/v0.18.1...v0.18.2) - 2025-10-10
- iaas-service-go v1.16.0 by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/669

## [v0.18.1](https://github.com/sacloud/autoscaler/compare/v0.18.0...v0.18.1) - 2025-10-08
- Handle scaling by default for non-Keep requests by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/666
- Change tag filtering to use TagsAndEqual by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/668

## [v0.18.0](https://github.com/sacloud/autoscaler/compare/v0.17.0...v0.18.0) - 2025-09-10
- go: bump google.golang.org/grpc from 1.74.2 to 1.75.0 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/639
- go: bump github.com/stretchr/testify from 1.10.0 to 1.11.1 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/643
- go: bump github.com/sacloud/api-client-go from 0.3.2 to 0.3.3 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/641
- go: bump google.golang.org/protobuf from 1.36.7 to 1.36.8 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/640
- feat: 台数維持機能 by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/663

## [v0.17.0](https://github.com/sacloud/autoscaler/compare/v0.16.2...v0.17.0) - 2025-08-14
- e2e: use terrafrm-provider-sakuracloud v2.27.0 by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/619
- golangci-lint v2 by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/624
- integration test and use go-version-file by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/626
- ci: bump docker/build-push-action from 5 to 6 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/613
- Fix YAML parsing issue by unmarshaling to string before checking by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/627
- go: bump github.com/prometheus/client_golang from 1.18.0 to 1.23.0 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/623
- go: bump google.golang.org/protobuf from 1.32.0 to 1.36.6 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/615
- sacloud/go-otelsetup v0.4.0 by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/628
- go: bump google.golang.org/protobuf from 1.36.6 to 1.36.7 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/630
- go: bump github.com/sacloud/go-otelsetup from 0.4.0 to 0.5.0 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/631
- go: bump github.com/sacloud/iaas-service-go from 1.9.2 to 1.13.0 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/629
- ci: bump actions/checkout from 4 to 5 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/635
- go: bump github.com/go-playground/validator/v10 from 10.23.0 to 10.27.0 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/634
- go: bump github.com/spf13/cobra from 1.8.0 to 1.9.1 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/633
- go: bump google.golang.org/grpc from 1.73.0 to 1.74.2 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/632
- use tagpr and goreleaser by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/636
- golangci-lint: errcheck and gosec by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/638

## [v0.16.2](https://github.com/sacloud/autoscaler/compare/v0.16.1...v0.16.2) - 2025-06-09
- go-playground/validator v10.15.4互換のcidrv4バリデーターを実装 by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/603
- Copyright 2025 by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/610
- iaas-api-go v1.15.0 by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/611
- update go-yaml to v1.18.0 by @repeatedly in https://github.com/sacloud/autoscaler/pull/612
- ci: bump goreleaser/goreleaser-action from 5 to 6 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/609

## [v0.16.1](https://github.com/sacloud/autoscaler/compare/v0.16.0...v0.16.1) - 2024-04-11
- downgrade: github.com/go-playground/validator to v10.15.4 by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/597

## [v0.16.0](https://github.com/sacloud/autoscaler/compare/v0.15.5...v0.16.0) - 2024-04-09
- go: bump github.com/prometheus/client_golang from 1.17.0 to 1.18.0 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/564
- go: bump github.com/c-robinson/iplib from 1.0.7 to 1.0.8 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/563
- go: bump google.golang.org/protobuf from 1.31.0 to 1.32.0 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/562
- go: bump google.golang.org/grpc from 1.60.0 to 1.60.1 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/561
- Instrumentation of traces with OpenTelemetry by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/565
- go: bump github.com/go-playground/validator/v10 from 10.16.0 to 10.17.0 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/567
- go: bump github.com/prometheus/common from 0.45.0 to 0.46.0 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/566
- OTel計装周りの修正 - 命名ルールの統一 by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/574
- Use otelsetup.InitWithOptions instead of otelsetup.Init by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/575
- update dependencies by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/587
- go: bump github.com/go-playground/validator/v10 from 10.17.0 to 10.19.0 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/585
- go: bump github.com/goccy/go-yaml from 1.11.2 to 1.11.3 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/573
- go: bump google.golang.org/grpc from 1.60.1 to 1.62.1 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/586

## [v0.15.5](https://github.com/sacloud/autoscaler/compare/v0.15.4...v0.15.5) - 2023-12-12
- go: bump github.com/stretchr/testify from 1.8.3 to 1.8.4 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/517
- go: bump github.com/go-playground/validator/v10 from 10.14.0 to 10.14.1 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/518
- go: bump github.com/sacloud/iaas-service-go from 1.9.0 to 1.9.1 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/519
- go: bump github.com/prometheus/client_golang from 1.15.1 to 1.16.0 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/521
- GPUプラン & AMDプラン by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/528
- Fix: memory size unit by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/530
- go: bump google.golang.org/protobuf from 1.30.0 to 1.31.0 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/524
- go: bump github.com/go-playground/validator/v10 from 10.14.1 to 10.15.0 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/527
- go 1.21 by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/532
- go-kit/log -> log/slog by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/533
- go: bump github.com/go-playground/validator/v10 from 10.15.0 to 10.16.0 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/548
- go: bump google.golang.org/grpc from 1.55.0 to 1.59.0 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/547
- ci: bump docker/login-action from 2 to 3 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/538
- ci: bump goreleaser/goreleaser-action from 4 to 5 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/536
- ci: bump docker/setup-buildx-action from 2 to 3 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/537
- ci: bump crazy-max/ghaction-import-gpg from 5 to 6 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/535
- ci: bump actions/checkout from 3 to 4 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/534
- go: bump github.com/c-robinson/iplib from 1.0.6 to 1.0.7 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/542
- go: bump github.com/goccy/go-yaml from 1.11.0 to 1.11.2 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/540
- go: bump github.com/sacloud/api-client-go from 0.2.8 to 0.2.9 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/541
- ci: bump docker/setup-qemu-action from 2 to 3 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/551
- ci: bump docker/metadata-action from 4 to 5 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/550
- ci: bump docker/build-push-action from 4 to 5 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/549
- go: bump github.com/prometheus/common from 0.44.0 to 0.45.0 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/552
- go: bump github.com/spf13/cobra from 1.7.0 to 1.8.0 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/553
- ci: bump actions/setup-go from 4 to 5 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/555
- go: bump github.com/sacloud/iaas-service-go from 1.9.2-0.20230808054001-efad52d748d4 to 1.9.2 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/556
- go: bump google.golang.org/grpc from 1.59.0 to 1.60.0 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/560

## [v0.15.4](https://github.com/sacloud/autoscaler/compare/v0.15.3...v0.15.4) - 2023-06-20
- Handlers: added missing parameters by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/522

## [v0.15.3](https://github.com/sacloud/autoscaler/compare/v0.15.2...v0.15.3) - 2023-05-25
- sacloud/iaas-service-go@v1.9.0 by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/515
- go: bump google.golang.org/grpc from 1.54.0 to 1.55.0 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/508
- go: bump github.com/prometheus/common from 0.42.0 to 0.44.0 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/513
- defaults.DefaultStatePollingTimeout -> 60 by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/516

## [v0.15.2](https://github.com/sacloud/autoscaler/compare/v0.15.1...v0.15.2) - 2023-05-09
- Set timeout for power operation by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/507
- go: bump github.com/goccy/go-yaml from 1.10.1 to 1.11.0 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/499
- go: bump github.com/sacloud/iaas-api-go from 1.9.1 to 1.10.0 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/503
- go: bump github.com/spf13/cobra from 1.6.1 to 1.7.0 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/500
- go: bump github.com/sacloud/iaas-service-go from 1.7.0 to 1.8.2 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/505
- go: bump github.com/prometheus/client_golang from 1.14.0 to 1.15.1 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/506

## [v0.15.1](https://github.com/sacloud/autoscaler/compare/v0.15.0...v0.15.1) - 2023-03-29
- inputs: directにおいてjobのステータスに応じて終了コードを出し分け by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/492
- go: bump github.com/sacloud/iaas-api-go from 1.8.3 to 1.9.0 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/493
- sacloud/iaas-api-go v1.9.1 by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/497
- go: bump github.com/sacloud/iaas-service-go from 1.6.1 to 1.7.0 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/496
- go: bump github.com/goccy/go-yaml from 1.10.0 to 1.10.1 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/498

## [v0.15.0](https://github.com/sacloud/autoscaler/compare/v0.14.1...v0.15.0) - 2023-03-22
- go: bump github.com/sacloud/iaas-service-go from 1.6.0 to 1.6.1 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/482
- go: bump github.com/goccy/go-yaml from 1.9.8 to 1.10.0 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/481
- go: bump github.com/stretchr/testify from 1.8.1 to 1.8.2 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/478
- ci: bump actions/setup-go from 2 to 3 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/477
- go: bump github.com/prometheus/common from 0.39.0 to 0.41.0 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/480
- go 1.20 by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/485
- go: bump google.golang.org/protobuf from 1.28.1 to 1.29.0 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/483
- go: bump github.com/prometheus/common from 0.41.0 to 0.42.0 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/484
- Added --sync parameter for synchronous up/down by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/491
- go: bump google.golang.org/grpc from 1.53.0 to 1.54.0 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/490
- go: bump google.golang.org/protobuf from 1.29.0 to 1.30.0 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/488
- ci: bump actions/setup-go from 3 to 4 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/487
- go: bump github.com/go-playground/validator/v10 from 10.11.2 to 10.12.0 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/489

## [v0.14.1](https://github.com/sacloud/autoscaler/compare/v0.14.0...v0.14.1) - 2023-02-28
- cooldownの基準時刻修正 by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/479

## [v0.14.0](https://github.com/sacloud/autoscaler/compare/v0.13.0...v0.14.0) - 2023-02-21
- docs: cooldownの基準変更 by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/474
- cooldownの基準として各リソースのModifiedAtを用いる by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/475

## [v0.13.0](https://github.com/sacloud/autoscaler/compare/v0.12.2...v0.13.0) - 2023-02-13
- cooldownパラメータをup/downで分離 by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/473
- go: bump google.golang.org/grpc from 1.52.0 to 1.53.0 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/471

## [v0.12.2](https://github.com/sacloud/autoscaler/compare/v0.12.1...v0.12.2) - 2023-02-09
- go: bump github.com/sacloud/iaas-service-go from 1.4.0 to 1.5.0 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/453
- e2e: is1a -> tk1b by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/456
- go: bump github.com/sacloud/iaas-api-go from 1.7.0 to 1.7.1 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/455
- go: bump github.com/goccy/go-yaml from 1.9.7 to 1.9.8 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/454
- copyright: 2023 by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/459
- go: bump github.com/c-robinson/iplib from 1.0.4 to 1.0.6 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/458
- go: bump google.golang.org/grpc from 1.51.0 to 1.52.0 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/460
- go: bump github.com/sacloud/iaas-service-go from 1.5.0 to 1.6.0 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/462
- go: bump github.com/sacloud/iaas-api-go from 1.8.0 to 1.8.1 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/463
- go: bump github.com/go-playground/validator/v10 from 10.11.1 to 10.11.2 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/467
- go: bump github.com/sacloud/iaas-api-go from 1.8.1 to 1.8.3 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/469
- ci: bump docker/build-push-action from 3 to 4 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/470
- ServerGroupにおけるゾーンの複数指定チェックバグの修正 by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/472

## [v0.12.1](https://github.com/sacloud/autoscaler/compare/v0.12.0...v0.12.1) - 2022-12-15
- go: bump github.com/prometheus/common from 0.37.0 to 0.38.0 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/447
- go: bump github.com/sacloud/packages-go from 0.0.6 to 0.0.7 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/448
- go: bump github.com/hashicorp/errwrap from 1.0.0 to 1.1.0 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/449
- ci: bump goreleaser/goreleaser-action from 3 to 4 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/450
- go: bump github.com/prometheus/common from 0.38.0 to 0.39.0 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/452
- go: bump github.com/sacloud/iaas-api-go from 1.6.2 to 1.7.0 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/451

## [v0.12.0](https://github.com/sacloud/autoscaler/compare/v0.11.2...v0.12.0) - 2022-12-06
- go: bump github.com/sacloud/api-client-go from 0.2.3 to 0.2.4 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/440
- go: bump google.golang.org/grpc from 1.50.1 to 1.51.0 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/441
- go: bump github.com/sacloud/iaas-api-go from 1.6.0 to 1.6.1 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/442
- go: bump github.com/c-robinson/iplib from 1.0.3 to 1.0.4 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/443
- validate.Error型の導入 by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/446
- go: bump github.com/goccy/go-yaml from 1.9.6 to 1.9.7 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/445
- go: bump github.com/sacloud/iaas-api-go from 1.6.1 to 1.6.2 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/444

## [v0.11.2](https://github.com/sacloud/autoscaler/compare/v0.11.1...v0.11.2) - 2022-11-15
- go: bump google.golang.org/grpc from 1.49.0 to 1.50.0 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/425
- go: bump github.com/sacloud/iaas-api-go from 1.4.1 to 1.5.1 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/429
- go: bump google.golang.org/grpc from 1.50.0 to 1.50.1 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/430
- go: bump github.com/spf13/cobra from 1.5.0 to 1.6.0 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/426
- go: bump github.com/stretchr/testify from 1.8.0 to 1.8.1 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/433
- go: bump github.com/spf13/cobra from 1.6.0 to 1.6.1 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/431
- go: bump github.com/sacloud/iaas-service-go from 1.3.2 to 1.4.0 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/432
- go: bump github.com/goccy/go-yaml from 1.9.5 to 1.9.6 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/435
- go: bump github.com/prometheus/client_golang from 1.13.0 to 1.14.0 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/437
- ELB: サーバグループに所属する実サーバの状況に応じてデタッチ処理をスキップ by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/438

## [v0.11.1](https://github.com/sacloud/autoscaler/compare/v0.11.0...v0.11.1) - 2022-10-06
- go 1.19 by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/418
- sacloud/makefile v0.0.7 by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/419
- go: bump github.com/sacloud/api-client-go from 0.2.1 to 0.2.2 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/420
- go: bump github.com/sacloud/packages-go from 0.0.5 to 0.0.6 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/421
- go: bump github.com/sacloud/iaas-service-go from 1.3.1 to 1.3.2 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/423
- ユーザーエージェントをカスタマイズするためのSAKURACLOUD_APPEND_USER_AGENT環境変数の導入 by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/424

## [v0.11.0](https://github.com/sacloud/autoscaler/compare/v0.10.1...v0.11.0) - 2022-09-30
- iaas-api-go v1.4.0 by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/417

## [v0.10.1](https://github.com/sacloud/autoscaler/compare/v0.10.0...v0.10.1) - 2022-09-28
- go: bump github.com/go-playground/validator/v10 from 10.11.0 to 10.11.1 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/412
- 正常にサーバー停止した場合にエラーを返さないように修正 by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/413
- infoログ追加: resource-name,desired-state-name by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/414
- セレクターでの検索エラーメッセージにゾーンを追加 by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/415

## [v0.10.0](https://github.com/sacloud/autoscaler/compare/v0.9.0...v0.10.0) - 2022-09-07
- グループタグによるホスト分散 by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/406
- 水平スケールでの名称フォーマット指定機能 by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/407
- HostNameFormatが空の場合のデフォルト値を文字列+数値を受け入れ可能なフォーマットにする by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/408

## [v0.9.0](https://github.com/sacloud/autoscaler/compare/v0.8.0...v0.9.0) - 2022-08-31
- IDを指定する項目でセレクタを指定可能に by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/403
- go: bump github.com/sacloud/iaas-api-go from 1.3.1 to 1.3.2 by @dependabot[bot] in https://github.com/sacloud/autoscaler/pull/402
- 水平スケールで複数ゾーンを利用可能に by @yamamoto-febc in https://github.com/sacloud/autoscaler/pull/404
