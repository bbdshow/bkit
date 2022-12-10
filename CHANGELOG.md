### v0.3.6
1. upgrade qezap client
### v0.3.5
1. Add SliceRemoveDuplicate
2. Fix:bug inviter.SetBase
### v0.3.4
1. Fix typ.ReqData struct
### v0.3.3
1. Add: typ.Resp for as common results struct
### v0.3.2
1. Fix: errc Multi.Msg format
### v0.3.1
1. Add: distance calc
### v0.3.0
1. Add: pomelo client as tcp test client
### v0.2.10
1. Add: httplib req withContext
2. delete api sign URL PATH func
### v0.2.9
1. inviteCode: Support [base length pad] customization
### v0.2.8
1. gen.str Add Substring  substr func
2. errc msg EN uppercase, add msg CN support
3. itime.CtxAfterSecDeadline
### v0.2.7
1. task runner: Deprecated AddTimeAfterFunc, replace to AddTickerTimeAfterFunc, AddOnceTimeAfterFunc
### v0.2.6
1. alert add telegram method implement
### v0.2.5
1. util package add httplib, http request client
### v0.2.4
1. ginutil: add http pprof handler
### v0.2.2
1. fixBug: xhttp Post sign url path
### v0.2.1
1. fix Req Id tag

### v0.2.0
1. add copier pkg, strcut mapping

### v0.1.10
1. no auth: httpCode 401 replace 403
### v0.1.9
1.fixBug: apiSign enable Path sign
### v0.1.8
1.fix go.mod
### v0.1.7
1. add ToInternalError
2. ad GetApiSignAccessKey
### v0.1.6
1. add iemail hermes
2. fixBug findSql
### v0.1.5
1. swag upgrade v1.7.1

### v0.1.4

### v0.1.3
1. open sign.DecodeHeaderVal sign.SignedValidTime
### v0.1.2
1.RandAlphaNumString, each cycle change rand.Source
2.add openId struct
### v0.1.1
1. fix runner, multi runners, use the same that context done
2. add runner example
3. runner.Server interface , argument ...option replace config

### v0.1.0
1. bkit init