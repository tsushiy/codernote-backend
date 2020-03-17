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
        "Title": "A. 積雪深差"
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
        "ProblemIDList": [
            "abc001_1",
            "abc001_2",
            "abc001_3",
            "abc001_4"
        ]
    }
]
```

### GET /notes

公開されているノートの一覧を取得します

#### Parameters

QueryString

- domain
- contestId
- problemId
- userName
- tag
- limit (can not exceed 1000)
- skip
- order: "-updated"

example: /notes?domain=atcoder&userName=tsushiy&tag=tag1&limit=100&skip=0&order=-updated

#### Response

```json
{
    "Count": 1,  // Total # of notes matched to query (domain, contestId, problemId, userName, tag)
    "Notes": [
        {
            "CreatedAt": "2020-03-15T11:38:48.04207Z",
            "UpdatedAt": "2020-03-15T11:41:43.371398Z",
            "Text": "sample text.",
            "Problem": {
                "No": 1,
                "Domain": "atcoder",
                "ProblemID": "abc001_1",
                "ContestID": "abc001",
                "Title": "A. 積雪深差"
            },
            "User": {
                "UserID": "fgCE5ZcTeOT8hmEmNnXvBb4mhEg1",
                "Name": "tsushiy",
                "CreatedAt": "2020-03-15T10:36:11.273197Z",
                "UpdatedAt": "2020-03-15T11:17:48.712348Z"
            },
            "Public": true
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

### GET /user/note

ログインしているユーザの指定された単一のノートを取得します。

#### Parameters

QueryString

- domain (required)
- contestId (required)
- problemId (required)

example: /user/note?domain=atcoder&contestId=abc001&problemId=abc001_1

#### Response

```json
{
    "CreatedAt": "2020-03-15T11:38:48.04207Z",
    "UpdatedAt": "2020-03-15T11:39:50.905595Z",
    "Text": "sample text.",
    "Problem": {
        "No": 1,
        "Domain": "atcoder",
        "ProblemID": "abc001_1",
        "ContestID": "abc001",
        "Title": "A. 積雪深差"
    },
    "User": {
        "UserID": "fgCE5ZcTeOT8hmEmNnXvBb4mhEg1",
        "Name": "tsushiy",
        "CreatedAt": "2020-03-15T10:36:11.273197Z",
        "UpdatedAt": "2020-03-15T11:17:48.712348Z"
    },
    "Public": true
}
```

### POST /user/note

ログインしているユーザで指定された単一のノートを投稿または更新します。

#### Parameters

Request Body

```json
{
    "Domain": "atcoder",     // required
    "ContestID": "abc001",   // required
    "ProblemID": "abc001_1", // required
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
- problemId
- tag
- limit (can not exceed 1000)
- skip
- order

example: /user/notes?domain=atcoder&userName=tsushiy&tag=tag1&limit=100&skip=0&order=-updated

#### Response

```json
{
    "Count": 1,  // Total # of notes matched to query (domain, contestId, problemId, tag)
    "Notes": [
        {
            "CreatedAt": "2020-03-15T11:38:48.04207Z",
            "UpdatedAt": "2020-03-15T11:41:43.371398Z",
            "Text": "sample text.",
            "Problem": {
                "No": 1,
                "Domain": "atcoder",
                "ProblemID": "abc001_1",
                "ContestID": "abc001",
                "Title": "A. 積雪深差"
            },
            "User": {
                "UserID": "fgCE5ZcTeOT8hmEmNnXvBb4mhEg1",
                "Name": "tsushiy",
                "CreatedAt": "2020-03-15T10:36:11.273197Z",
                "UpdatedAt": "2020-03-15T11:17:48.712348Z"
            },
            "Public": true
        }
    ]
}
```

### GET /user/note/tag

ログインしているユーザの指定されたノートのタグ一覧を取得します。

#### Parameters

QueryString

- domain (required)
- contestId (required)
- problemId (required)

example: /user/note/tag?domain=atcoder&contestId=abc001&problemId=abc001_1

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

### POST /user/note/tag

ログインしているユーザの指定されたノートにタグを追加します。

#### Parameters

Request Body

```json
{
    "Domain": "atcoder",     // required
    "ContestID": "abc001",   // required
    "ProblemID": "abc001_1", // required
    "Tag": "tag1",           // required
}
```

#### Response

* 200: OK

### DELETE /user/note/tag

ログインしているユーザの指定されたノートからタグを削除します。

#### Parameters

Request Body

```json
{
    "Domain": "atcoder",     // required
    "ContestID": "abc001",   // required
    "ProblemID": "abc001_1", // required
    "Tag": "tag1",           // required
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
Contest {
    No               int
    Domain           string
    ContestID        string
    Title            string
    StartTimeSeconds int
    DurationSeconds  int
    ProblemIDList    []string
}
```

```
Problem {
    No        int
    Domain    string
    ProblemID string
    ContestID string
    Title     string
}
```

```
Note {
    No        int
    CreatedAt string (RFC 3339)
    UpdatedAt string (RFC 3339)
    Text      string
    ProblemNo int
    Problem   Problem
    UserNo    int
    User      User
    Public    bool
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
    NoteNo int
    TagNo  int
}
```