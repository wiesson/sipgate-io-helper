start ngrok

```
ngrok http 3000 --region=eu
```

run io-helper:

```
go run main.go your_email@domain.tld web_password
```

Output:

```
Found ngrok url https://65f01afc.eu.ngrok.io
2017/01/19 12:42:42 map[direction:[out] from:[4920387844349] to:[4921163555757] callId:[2327720023710239030] user[]:[Peterle Drobusch-Xjg] event:[newCall]]
2017/01/19 12:42:50 map[from:[4920387844349] direction:[out] event:[hangup] callId:[2327720023710239030] cause:[cancel] to:[4921163555757]]
```
