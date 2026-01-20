# Judge-Service
DockerとGoのGinが必要です。

# 概要
これはPathlyの課題題材における採点のランタイムを隔離するサンドボックスとして作成したものです。
現状、Cのランタイムのみ対応可能で他もDocker in Docker(他言語のランタイムを走らせればいいので)のため拡張性もあります。

# 仕組み

逆方向もレスポンスだけのため省略します。
となっております。
<img width="1001" height="701" alt="ジャッジサービス drawio" src="https://github.com/user-attachments/assets/8ed99262-e1cf-4744-b32e-fda318a3d0e3" />

