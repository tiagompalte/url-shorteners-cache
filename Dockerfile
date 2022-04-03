FROM golang:1.18 as build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
COPY *.go ./

RUN go mod download
RUN go build -o main main.go

FROM gcr.io/distroless/base-debian10

WORKDIR /

COPY --from=build /app/main /
COPY .env.docker .env

EXPOSE 8000

USER nonroot:nonroot

ENTRYPOINT ["/main"]