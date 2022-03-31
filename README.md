# tele-go

# Telegram Bot for OpenWRT router



### Compilation binary-file for MIPS:
>    GOOS=linux GOARCH=mipsle GOMIPS=softfloat go build -trimpath -ldflags="-s -w" -o telego