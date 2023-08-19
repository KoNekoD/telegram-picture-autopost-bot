build:
	go build -ldflags "-s -w" . && \
    GOOS=windows GOARCH=386 go build -ldflags "-s -w" -o telegramPicAutopostBot.exe && \
    GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o telegramPicAutopostBot64.exe