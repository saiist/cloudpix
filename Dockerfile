FROM golang:1.20 as build
WORKDIR /cloudpix

# Copy dependencies list
COPY go.mod go.sum ./

# Build with optional lambda.norpc tag
COPY cmd/ ./cmd/
RUN go build -tags lambda.norpc -o main ./cmd/upload/main.go

# Copy artifacts to a clean image
FROM public.ecr.aws/lambda/provided:al2023
COPY --from=build /cloudpix/main ./main
ENTRYPOINT [ "./main" ]