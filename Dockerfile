FROM golang:alpine

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

ARG TARGETOS
ARG TARGETARCH
ARG GOARCH=$TARGETARCH \
    GOOS=$TARGETOS

RUN CGO_ENABLED=0 GOOS=linux go build -v -a -installsuffix cgo -o main cmd/todos-app/main.go

RUN cp /app/main /bin/main

RUN adduser -D appuser
USER appuser

EXPOSE 8080

CMD ["/bin/main", "s"]
