# goj

勉強として作っているAtCoderの補助ツールです。
(プログラミング初心者が作ったツールなので品質は察してください)

## 使い方
このツールは、
```
.
├── abc174
│   ├── abc174_c.cpp
│   ├── goj.toml
│   └── test_abc174_c
│       ├── sample-1.in
│       ├── sample-1.out
│       ├── sample-2.in
│       ├── sample-2.out
│       ├── sample-3.in
│       └── sample-3.out
└── abc175
    ├── abc175_a.cpp
    ├── abc175_b.cpp
    ├── goj.toml
    ├── test_abc175_a
    │   ├── sample-1.in
    │   ├── sample-1.out
    │   ├── sample-2.in
    │   ├── sample-2.out
    │   ├── sample-3.in
    │   └── sample-3.out
    └── test_abc175_b
        ├── sample-1.in
        ├── sample-1.out
        ├── sample-2.in
        ├── sample-2.out
        ├── sample-3.in
        ├── sample-3.out
        ├── sample-4.in
        └── sample-4.out
```
のように1つのコンテストに1つのディレクトリを割り当て、
各問題は各コンテストのディレクトリの直下に置くような構造を用いた時に
使いやすくなるように設計されています。
また、~/.config/goj/config.tomlに設定ファイルがありデフォルト言語等を設定出来ます。


## ログイン
```
$ goj login
```
でユーザー名とパスワードを聞かれるので入力してください。
cookiejarが~/.cachd/goj/cookiejarに保存されます。
コンテスト本番時以外はログインは不要です。


## テストケースのダウンロード
```
$ goj download abc175/abc175_a
```
で問題のサンプルケースをダウンロードするとともに問題のテンプレートファイルを作成します。
問題名を省略するとコンテストの問題すべてをダウンロードします。
コンテスト名も省略するとカレントディレクトリの名前をコンテスト名と見なします。


## テスト
```
$ goj test abc175_b
```
で問題abc175_bのテストを行います。
問題をabc175_bをテストしたい時、与える引数が問題名のsuffixに一致していれば大丈夫です。
つまり最後に編集したのがB問題であれば`goj test b`で構いません。
問題名を省略すると最後に編集されたファイルの名前を問題名と見なします。
`-command <command>`でテストするコマンドを指定できます。この場合問題名は必須です。


## 提出
```
$ goj submit abc175_b
```
でabc175_b.cpp(デフォルトの場合)をテストしたのち問題abc175_bに提出します。
問題名を省略すると最後に編集されたファイルの名前を提出する問題と見なします。
`-f`でテストをスキップ出来ます。


## demo
![demo](demo.gif)
