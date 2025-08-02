# ビルドステージ
FROM golang:1.24-alpine AS builder

WORKDIR /app

# git と zip をインストール（zipはここで入れておく）
RUN apk add --no-cache git zip

COPY go.mod go.sum ./

ENV GOPROXY=direct

RUN go mod download

COPY . .

# Linux用バイナリをビルド
RUN GOOS=linux GOARCH=amd64 go build -o bootstrap main.go

# バイナリをzip化（同じビルドステージ内）
RUN zip function.zip bootstrap

# 実行用ステージ（Lambda用ランタイムイメージ）
FROM public.ecr.aws/lambda/go:1

# バイナリをコピー
COPY --from=builder /app/bootstrap ${LAMBDA_RUNTIME_DIR}/bootstrap

# （必要ならzipファイルも取り出すためにここにコピーすることも可能）
COPY --from=builder /app/function.zip /function.zip

CMD ["bootstrap"]
