FROM golang:1.25.6-alpine3.22 AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -o seed-admin ./cmd/seed-admin

FROM alpine:3.22

RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    libreoffice \
    fontconfig \
    font-carlito \
    font-liberation \
    font-dejavu

COPY fonts.conf /etc/fonts/local.conf

RUN fc-cache -f \
    && fc-match Calibri | grep -qi Carlito \
    && fc-match Arial | grep -qi "Liberation Sans" \
    && fc-match "Times New Roman" | grep -qi "Liberation Serif"
ENV TZ=Asia/Jakarta

WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/seed-admin .

EXPOSE 8080

CMD ["./main"]
