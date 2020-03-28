# Codernote Backend

Frontend: [codernote-frontend](https://github.com/tsushiy/codernote-frontend)

## Develop on your local

### Run API Server

```sh
go build
./codernote-backend
```

### Run Crawler

```sh
cd crawler
go run cmd/main.go
```

ルート以下のAPIサーバと`crawler/`以下のCrawlerは別モジュールになっています。  
CrawlerはAPIサーバ側のdbパッケージに依存しているので、バージョン管理に注意してください。  
例えばDBの構成を変更する場合には、`crawler/go.mod`に以下のように追記してローカルパッケージを用いる、といった対応をしてください。

```
replace github.com/tsushiy/codernote-backend => ../
```

## Crawler

以下のAPIから取得したデータを同じ形式にしてデータベースに格納します。

- [AtCoderProblems API](https://github.com/kenkoooo/AtCoderProblems)
- [Codeforces API](https://codeforces.com/apiHelp)
- [AOJ API](http://developers.u-aizu.ac.jp/index)
- [yukicoder API](https://petstore.swagger.io/?url=https://yukicoder.me/api/swagger.yaml)
- [LeetCode API](https://leetcode.com/api/problems/algorithms/)

## API Server

apiv1.codernote.tsushiy.com で呼べますが、codernote-frontend 以外から呼ばれることはあまり想定していません。

Crawlerで取得しておいたデータを用いて問題のバリデーションを行います。

認証が必要なAPIでは、Firebase Authenticationで得られたJWTの検証を行います。

## Non-Auth API

### GET /problems

問題の一覧を取得します

#### Parameters

QueryString

- domain: "atcoder"

example: /problems?domain=atcoder

#### Response

```json
[
    {
        "No": 1,
        "Domain": "atcoder",
        "ProblemID": "abc001_1",
        "ContestID": "abc001",
        "Title": "A. 積雪深差",
        "Slug":"",
        "FrontendID":"",
        "Difficulty":"194.98182678222656"
    }
]
```

### GET /contests

コンテストの一覧を取得します

#### Parameters

QueryString

- domain: "atcoder"
- order: "-started", "started"

example: /contests?domain=atcoder&order=-started

#### Response

```json
[
    {
        "No": 5,
        "Domain": "atcoder",
        "ContestID": "abc001",
        "Title": "AtCoder Beginner Contest 001",
        "StartTimeSeconds": 1381579200,
        "DurationSeconds": 7200,
        "ProblemNoList": [
            1,
            2,
            3,
            4
        ]
    }
]
```

### GET /note

公開されている単一のノートを取得します

#### Parameters

QueryString

- noteId (required)

example: /note?noteId=74b3ea1e-b296-4d62-bb9a-81fa5c39dd31

#### Response

```json
{
    "ID": "74b3ea1e-b296-4d62-bb9a-81fa5c39dd31",
    "CreatedAt": "2020-03-15T11:38:48.04207Z",
    "UpdatedAt": "2020-03-15T11:41:43.371398Z",
    "Text": "sample text.",
    "Problem": {
        "No": 1,
        "Domain": "atcoder",
        "ProblemID": "abc001_1",
        "ContestID": "abc001",
        "Title": "A. 積雪深差",
        "Slug":"",
        "FrontendID":"",
        "Difficulty":"194.98182678222656"
    },
    "User": {
        "UserID": "fgCE5ZcTeOT8hmEmNnXvBb4mhEg1",
        "Name": "tsushiy",
        "CreatedAt": "2020-03-15T10:36:11.273197Z",
        "UpdatedAt": "2020-03-15T11:17:48.712348Z"
    },
    "Public": 2
}
```

### GET /notes

公開されているノートの一覧を取得します

#### Parameters

QueryString

- domain
- problemNo
- contestId
- userName
- tag
- limit (can not exceed 1000)
- skip
- order: "-updated"

example: /notes?domain=atcoder&userName=tsushiy&tag=tag1&limit=100&skip=0&order=-updated

#### Response

```json
{
    "Count": 1,  // Total # of notes matched to the query (domain, problemNo, contestId, userName, tag)
    "Notes": [
        {
            "ID": "74b3ea1e-b296-4d62-bb9a-81fa5c39dd31",
            "CreatedAt": "2020-03-15T11:38:48.04207Z",
            "UpdatedAt": "2020-03-15T11:41:43.371398Z",
            "Text": "sample text.",
            "Problem": {
                "No": 1,
                "Domain": "atcoder",
                "ProblemID": "abc001_1",
                "ContestID": "abc001",
                "Title": "A. 積雪深差",
                "Slug":"",
                "FrontendID":"",
                "Difficulty":"194.98182678222656"
            },
            "User": {
                "UserID": "fgCE5ZcTeOT8hmEmNnXvBb4mhEg1",
                "Name": "tsushiy",
                "CreatedAt": "2020-03-15T10:36:11.273197Z",
                "UpdatedAt": "2020-03-15T11:17:48.712348Z"
            },
            "Public": 2
        }
    ]
}
```

## Auth API

A JWT must be included in the header of the request.

```json
{
    "Authorization": "Bearer {JWT}"
}
```

### POST /login

ログインします。  
初回ログイン時にはランダムなユーザネームで新しく登録されます。

#### Parameters

#### Response

```json
{
    "UserID": "fgCE5ZcTeOT8hmEmNnXvBb4mhEg1",
    "Name": "tsushiy",
    "CreatedAt": "2020-03-15T08:36:01.629775Z",
    "UpdatedAt": "2020-03-15T08:37:57.083458Z"
}
```

### POST /user/name

ユーザ名を変更します。  
ユーザ名は3文字から30文字の英数字である必要があります。

#### Parameters

Request Body

```json
{
    "Name": "tsushiy" // must be between 3 and 30 alphanumeric characters
}
```

#### Response

```json
{
    "UserID": "fgCE5ZcTeOT8hmEmNnXvBb4mhEg1",
    "Name": "tsushiy",
    "CreatedAt": "2020-03-15T08:36:01.629775Z",
    "UpdatedAt": "2020-03-15T08:41:02.216072Z"
}
```

### GET /user/setting

ユーザ設定を取得します。

#### Parameters

#### Response

```json
{
    "AtCoderID":    "",
    "CodeforcesID": "",
    "YukicoderID":  "",
    "AOJID":        "",
    "LeetCodeID":   "",
}
```

### POST /user/setting

ユーザ設定を変更します。

#### Parameters

Request Body

```json
{
    "AtCoderID":    "",
    "CodeforcesID": "",
    "YukicoderID":  "",
    "AOJID":        "",
    "LeetCodeID":   "",
}
```

#### Response

```json
{
    "AtCoderID":    "",
    "CodeforcesID": "",
    "YukicoderID":  "",
    "AOJID":        "",
    "LeetCodeID":   "",
}
```

### GET /user/note

公開されている単一のノートを取得します  
ログインしているユーザの作成したノートであれば、公開されていなくても取得します。

GET /note とほぼ同じです。ログインした状態であれば基本的にこっちを使えばいいです。

#### Parameters

QueryString

- noteId (required)

example: /user/note?noteId=74b3ea1e-b296-4d62-bb9a-81fa5c39dd31

#### Response

```json
{
    "ID": "74b3ea1e-b296-4d62-bb9a-81fa5c39dd31",
    "CreatedAt": "2020-03-15T11:38:48.04207Z",
    "UpdatedAt": "2020-03-15T11:41:43.371398Z",
    "Text": "sample text.",
    "Problem": {
        "No": 1,
        "Domain": "atcoder",
        "ProblemID": "abc001_1",
        "ContestID": "abc001",
        "Title": "A. 積雪深差",
        "Slug":"",
        "FrontendID":"",
        "Difficulty":"194.98182678222656"
    },
    "User": {
        "UserID": "fgCE5ZcTeOT8hmEmNnXvBb4mhEg1",
        "Name": "tsushiy",
        "CreatedAt": "2020-03-15T10:36:11.273197Z",
        "UpdatedAt": "2020-03-15T11:17:48.712348Z"
    },
    "Public": 2
}
```

### GET /user/note/{ProblemNo}

ログインしているユーザの指定された単一のノートを取得します。

#### Parameters

Path

- ProblemNo (required)

example: /user/note/1

#### Response

```json
{
    "ID": "74b3ea1e-b296-4d62-bb9a-81fa5c39dd31",
    "CreatedAt": "2020-03-15T11:38:48.04207Z",
    "UpdatedAt": "2020-03-15T11:39:50.905595Z",
    "Text": "sample text.",
    "Problem": {
        "No": 1,
        "Domain": "atcoder",
        "ProblemID": "abc001_1",
        "ContestID": "abc001",
        "Title": "A. 積雪深差",
        "Slug":"",
        "FrontendID":"",
        "Difficulty":"194.98182678222656"
    },
    "User": {
        "UserID": "fgCE5ZcTeOT8hmEmNnXvBb4mhEg1",
        "Name": "tsushiy",
        "CreatedAt": "2020-03-15T10:36:11.273197Z",
        "UpdatedAt": "2020-03-15T11:17:48.712348Z"
    },
    "Public": 2  // 1 if private, otherwise 2
}
```

### POST /user/note/{ProblemNo}

ログインしているユーザで指定された単一のノートを投稿または更新します。

#### Parameters

Path

- ProblemNo (required)

Request Body

```json
{
    "Text": "sample text.",  // required, must not be empty
    "Public": true           // false if empty
}
```

#### Response

* 200: OK

### GET /user/notes

ログインしているユーザのノートの一覧を取得します。

#### Parameters

QueryString

- domain
- contestId
- tag
- limit (can not exceed 1000)
- skip
- order

example: /user/notes?domain=atcoder&tag=tag1&limit=100&skip=0&order=-updated

#### Response

```json
{
    "Count": 1,  // Total # of notes matched to the query (domain, contestId, tag)
    "Notes": [
        {
            "ID": "74b3ea1e-b296-4d62-bb9a-81fa5c39dd31",
            "CreatedAt": "2020-03-15T11:38:48.04207Z",
            "UpdatedAt": "2020-03-15T11:41:43.371398Z",
            "Text": "sample text.",
            "Problem": {
                "No": 1,
                "Domain": "atcoder",
                "ProblemID": "abc001_1",
                "ContestID": "abc001",
                "Title": "A. 積雪深差",
                "Slug":"",
                "FrontendID":"",
                "Difficulty":"194.98182678222656"
            },
            "User": {
                "UserID": "fgCE5ZcTeOT8hmEmNnXvBb4mhEg1",
                "Name": "tsushiy",
                "CreatedAt": "2020-03-15T10:36:11.273197Z",
                "UpdatedAt": "2020-03-15T11:17:48.712348Z"
            },
            "Public": 2
        }
    ]
}
```

### GET /user/note/{ProblemNo}/tag

ログインしているユーザの指定されたノートのタグ一覧を取得します。

#### Parameters

Path

- ProblemNo (required)

example: /user/note/1/tag

#### Response

```json
{
    "Tags": [
        "tag1",
        "tag2",
        "tag3",
    ]
}
```

### POST /user/note/{ProblemNo}/tag

ログインしているユーザの指定されたノートにタグを追加します。

#### Parameters

Path

- ProblemNo (required)

Request Body

```json
{
    "Tag": "tag1",  // required
}
```

#### Response

* 200: OK

### DELETE /user/note/{ProblemNo}/tag

ログインしているユーザの指定されたノートからタグを削除します。

#### Parameters

Path

- ProblemNo (required)

Request Body

```json
{
    "Tag": "tag1",  // required
}
```

#### Response

* 200: OK

## Schemas

```
User {
    No        int
    UserID    string
    Name      string
    CreatedAt string (RFC 3339)
    UpdatedAt string (RFC 3339)
}
```

```
UserDetail struct {
    UserID       string
    AtCoderID    string
    CodeforcesID string
    YukicoderID  string
    AOJID        string
    LeetCodeID   string
}
```

```
Contest {
    No               int
    Domain           string
    ContestID        string
    Title            string
    StartTimeSeconds int
    DurationSeconds  int
    Rated            string
    ProblemNoList    []int
}
```

```
Problem {
    No         int
    Domain     string
    ProblemID  string
    ContestID  string
    Title      string
    Slug       string
    FrontendID string
    Difficulty string
}
```

```
Note {
    ID        string
    CreatedAt string (RFC 3339)
    UpdatedAt string (RFC 3339)
    Text      string
    ProblemNo int
    Problem   Problem
    UserNo    int
    User      User
    Public    int
}
```

```
Tag {
    No  int
    Key string
}
```

```
TagMap {
    No     int
    NoteID string
    TagNo  int
}
```
