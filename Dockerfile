FROM golang:1.22 AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -o /rehearsal ./cmd/server

FROM gcr.io/distroless/static
WORKDIR /app
COPY --from=build /rehearsal /rehearsal
COPY data/ data/
COPY web/ web/
COPY prompts/ prompts/
COPY spark/cache/ spark/cache/
CMD ["/rehearsal"]
