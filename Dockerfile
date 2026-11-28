FROM golang:1.25-alpine AS build

WORKDIR /src

# Lets a server where proxy.golang.org is blocked pick another module proxy; empty keeps Go's default.
ARG GOPROXY

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /out/ ./cmd/api ./cmd/coachctl

FROM alpine:3.20

RUN apk add --no-cache ca-certificates

COPY --from=build /out/ /usr/local/bin/

ENTRYPOINT ["api"]
