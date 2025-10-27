FROM golang:alpine AS builder

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

COPY . .

RUN --mount=type=cache,target=/root/.cache/go-build \
  go build -ldflags="-s -w" -o /main .

FROM scratch

COPY --from=builder /main /main

EXPOSE 80

ENTRYPOINT ["/main"]
