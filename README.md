# A source code converter
Transform `go` source code to `ts`(Typescript) code with best effort.

# Enum conversion
Usage:
```sh
go run ./ some-file.go
# or read from clipboard(macOS)
pbpaste|go run ./
```

Example `go` source:
```go
type LogHint string

const (
	LogHint_Empty                 LogHint = ""
	LogHint_MissingOver2Days      LogHint = "missing_over_2_days"
	LogHint_ResolvingWithin1Day   LogHint = "resolving_within_1_day"
	LogHint_ResolvingWithin8Hours LogHint = "resolving_within_8_hours"
	LogHint_WaitingResolve        LogHint = "waiting_resolve"
	LogHint_LogCompleteNeedManual LogHint = "log_complete_need_manual"
)
```

`ts` output:
```ts
// auto generated 
enum LogHint {
    Empty = "",
    MissingOver2Days = "missing_over_2_days",
    ResolvingWithin1Day = "resolving_within_1_day",
    ResolvingWithin8Hours = "resolving_within_8_hours",
    WaitingResolve = "waiting_resolve",
    LogCompleteNeedManual = "log_complete_need_manual",
}
const logHintValues:LogHint[] = [LogHint.Empty,LogHint.MissingOver2Days,LogHint.ResolvingWithin1Day,LogHint.ResolvingWithin8Hours,LogHint.WaitingResolve,LogHint.LogCompleteNeedManual]
const logHintMapping:Record<LogHint,string> = {
    [LogHint.Empty]: "",
    [LogHint.MissingOver2Days]: "",
    [LogHint.ResolvingWithin1Day]: "",
    [LogHint.ResolvingWithin8Hours]: "",
    [LogHint.WaitingResolve]: "",
    [LogHint.LogCompleteNeedManual]: "",
}
```

Compare that with some other converter:
```ts
type LogHint = string;
const LogHint_Empty: LogHint = "";
const LogHint_MissingOver2Days: LogHint = "missing_over_2_days";
const LogHint_ResolvingWithin1Day: LogHint = "resolving_within_1_day";
const LogHint_ResolvingWithin8Hours: LogHint = "resolving_within_8_hours";
const LogHint_WaitingResolve: LogHint = "waiting_resolve";
const LogHint_LogCompleteNeedManual: LogHint = "log_complete_need_manual";
```