FROM golang:1.24-alpine as build

WORKDIR /go/src/api
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 go build -o /go/bin/app ./app

FROM gcr.io/distroless/static-debian11 AS runtime
COPY --from=build /go/bin/app /app
COPY --from=build /go/src/api/credentials.json /credentials.json
ENTRYPOINT ["/app"]
