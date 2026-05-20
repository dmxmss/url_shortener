FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/redirect-service ./redirect-service/cmd

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/redirect-service /redirect-service
USER nonroot:nonroot
EXPOSE 8081
ENTRYPOINT ["/redirect-service"]

