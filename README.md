# Codernote Backend

Frontend: [codernote-frontend](https://github.com/tsushiy/codernote-frontend)

## Develop on your local

ルート以下のAPIサーバーと`crawler/`以下のCrawlerは別モジュールになっています。  
CrawlerはAPIサーバー側のdbパッケージに依存しているので、バージョン管理に注意してください。

### Run API Server

```sh
go build
./codernote-backend
```

### Run Crawler

```sh
go run ./crawler/cmd/main.go
```

# API

## Non-Auth API

### GET /problems?domain={domain}

問題の一覧を取得します

#### Parameters

Query

- domain: "atcoder"

#### Response

```json
[
    {
        "Domain": "atcoder",
        "ProblemID": "abc001_1",
        "ContestID": "abc001",
        "Title": "A. 積雪深差"
    }
]
```

### GET /contests?domain={domain}

コンテストの一覧を取得します

#### Parameters

Query

- domain: "atcoder"

#### Response

```json
[
    {
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

Request Body

```json
{
    "Domain": "atcoder",
    "ContestID": "abc001",
    "ProblemID": "abc001_1",
    "UserName": "tsushiy",
    "Tag": "sample tag",
    "Limit": 100, // cannot exceed 1000
    "Skip": 0,
    "Order": "-updated"
}
```

#### Response

```json
{
    "Count": 1,
    "Notes": [
        {
            "CreatedAt": "2020-03-15T11:38:48.04207Z",
            "UpdatedAt": "2020-03-15T11:41:43.371398Z",
            "Text": "sample text.",
            "Problem": {
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

### GET /login

ログインします。

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

### POST /user/changename

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

Request Body

```json
{
    "Domain": "atcoder",    // required
    "ContestID": "abc001",  // required
    "ProblemID": "abc001_1" // required
}
```

#### Response

```json
{
    "CreatedAt": "2020-03-15T11:38:48.04207Z",
    "UpdatedAt": "2020-03-15T11:39:50.905595Z",
    "Text": "sample text.",
    "Problem": {
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

Request Body

```json
{
    "Domain": "atcoder",
    "ContestID": "abc001",
    "ProblemID": "abc001_1",
    "Tag": "sample tag",
    "Limit": 100, // cannot exceed 1000
    "Skip": 0,
    "Order": "-updated"
}
```

#### Response

```json
{
    "Count": 1,
    "Notes": [
        {
            "CreatedAt": "2020-03-15T11:38:48.04207Z",
            "UpdatedAt": "2020-03-15T11:41:43.371398Z",
            "Text": "sample text.",
            "Problem": {
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

Request Body

```json
{
    "Domain": "atcoder",     // required
    "ContestID": "abc001",   // required
    "ProblemID": "abc001_1", // required
}
```

#### Response

```json
{
    "Tags": [
        "sample tag",
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
    "Tag": "sample tag",     // required
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
    "Tag": "sample tag",     // required
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