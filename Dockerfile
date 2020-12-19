FROM golang:alpine as gobuilder
ARG TARGETOS
ARG TARGETARCH

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /gosrc
COPY . .
RUN go build -o ascii -v cmd/main.go

# TODO: only supporting linux-based builds rn
FROM SCRATCH
WORKDIR /app
COPY --from=gobuilder /gosrc/ascii .
EXPOSE 8080

CMD ["/app/ascii"]