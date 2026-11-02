FROM golang:1.22-alpine AS build
WORKDIR /app/
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY cmd/ cmd
COPY internal/ internal
COPY pkg/ pkg
RUN CGO_ENABLED=0 GOOS=linux go build -installsuffix cgo -tags=nomsgpack -a -o bin/capturemapsapi cmd/capturemapsapi/main.go

FROM alpine:3.20
LABEL NAME="capture-maps-api"
RUN adduser -D -u 10001 app
WORKDIR /app
EXPOSE 8090
COPY configs/config.toml.template configs/config.toml
COPY --from=build /app/bin/capturemapsapi ./
RUN chown -R app:app /app
USER app
CMD ["./capturemapsapi"]
