FROM golang:latest AS Build
LABEL authors="mehedi"

WORKDIR /app

COPY go.mod ./
RUN go mod tidy

COPY . ./

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o go-stream

FROM alpine AS production
RUN apk update && apk add ffmpeg

COPY --from=Build /app/go-stream .

EXPOSE 9090

CMD ["./go-stream"]