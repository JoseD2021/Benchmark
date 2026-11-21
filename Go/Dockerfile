FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /api .

FROM scratch
COPY --from=build /api /api
EXPOSE 5001
ENTRYPOINT ["/api"]
