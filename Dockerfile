FROM golang:1.26.4-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/jobscout-api ./cmd/api

FROM alpine:3.22

RUN addgroup -S jobscout && adduser -S jobscout -G jobscout

WORKDIR /app

COPY --from=build /out/jobscout-api /app/jobscout-api

USER jobscout

EXPOSE 8080

ENTRYPOINT ["/app/jobscout-api"]
