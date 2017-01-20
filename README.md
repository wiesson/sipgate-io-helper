## sipgate io logger

### Installation

```
go get github.com/wiesson/sipgate-io-helper
cd $GOPATH/src/github.com/wiesson/sipgate-io-helper
go install
```

or download the binary (coming soon)

### Usage

1. Start Ngrok

```
ngrok http 3000 --region=eu
```

2. Start sipgate.io helper

```
sipgate-io-helper -email=your_email@domain.tld -password=12345678 -env=live
```

### Request token

```
sipgate-io-helper -email=your_email@domain.tld -password=12345678 -env=live -token=true
```

Sample Output:

```
Found ngrok url https://65f01afc.eu.ngrok.io
2017/01/19 12:42:42 map[direction:[out] from:[4920387844349] to:[4921163555757] callId:[2327720023710239030] user[]:[Peterle Drobusch-Xjg] event:[newCall]]
2017/01/19 12:42:50 map[from:[4920387844349] direction:[out] event:[hangup] callId:[2327720023710239030] cause:[cancel] to:[4921163555757]]
```
