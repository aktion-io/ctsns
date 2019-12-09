FROM golang:1.13 as build-env

WORKDIR /ctsns

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w"

FROM gcr.io/distroless/static
COPY --from=build-env /ctsns/ctsns /
CMD ["/ctsns"]
