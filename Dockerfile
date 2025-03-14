FROM golang:1.20 as build
WORKDIR /cloudpix

# Copy dependencies list
COPY go.mod go.sum ./

# BUILD_PATHをARGとして定義（デフォルト値あり）
ARG BUILD_PATH=./cmd/upload/main.go

# Copy source files
COPY cmd/ ./cmd/
COPY config/ ./config/
COPY internal/ ./internal/

# Build with optional lambda.norpc tag
RUN echo "Building from $BUILD_PATH" && \
    go build -tags lambda.norpc -o main $BUILD_PATH

# Copy artifacts to a clean image
FROM public.ecr.aws/lambda/provided:al2023
COPY --from=build /cloudpix/main ./main
ENTRYPOINT [ "./main" ]