FROM golang:1.16-alpine3.13 as builder
LABEL maintainer="Sudharshan"
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

######## Start a new stage from scratch #######
FROM alpine:3.13
WORKDIR /root/
RUN apk add --no-cache tzdata
COPY --from=builder /app/main .
COPY fb-creds.json .
EXPOSE 8000
CMD ["sh", "-c", "'sleep 20 && ./main'"]
