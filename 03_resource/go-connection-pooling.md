AWS SDK Connection Pooling (Go)

The AWS SDK for Go v2 uses the standard Go http.Client which has built-in connection pooling. You need to configure it properly:

package main

import (
"context"
"net"
"net/http"
"time"

      "github.com/aws/aws-sdk-go-v2/config"
      "github.com/aws/aws-sdk-go-v2/service/s3"

)

func NewOptimizedS3Client(ctx context.Context) (\*s3.Client, error) {
// Custom HTTP client with optimized connection pooling
httpClient := &http.Client{
Transport: &http.Transport{
// Maximum idle connections across all hosts
MaxIdleConns: 200,

              // Maximum idle connections per host (S3 endpoint)
              MaxIdleConnsPerHost: 200,

              // Maximum total connections per host
              MaxConnsPerHost: 200,

              // How long to keep idle connections alive
              IdleConnTimeout: 90 * time.Second,

              // Timeout for establishing new connections
              DialContext: (&net.Dialer{
                  Timeout:   30 * time.Second,
                  KeepAlive: 30 * time.Second,
              }).DialContext,

              // Disable compression (S3 data is often already compressed)
              DisableCompression: true,

              // Timeouts
              TLSHandshakeTimeout:   10 * time.Second,
              ResponseHeaderTimeout: 10 * time.Second,
              ExpectContinueTimeout: 1 * time.Second,
          },
          Timeout: 5 * time.Minute, // Overall request timeout
      }

      // Load AWS config with custom HTTP client
      cfg, err := config.LoadDefaultConfig(ctx,
          config.WithHTTPClient(httpClient),
          config.WithRegion("eu-west-2"),
      )
      if err != nil {
          return nil, err
      }

      // Create S3 client
      return s3.NewFromConfig(cfg), nil

}

Key Configuration Parameters:

MaxIdleConnsPerHost: 200

- Most important setting!
- Keeps 200 connections open and ready to S3
- Default is only 2 (way too low!)
- Should match or exceed your worker count

MaxConnsPerHost: 200

- Total concurrent connections allowed
- Prevents creating too many connections

IdleConnTimeout: 90s

- How long to keep unused connections alive
- Longer is better for sustained operations

DisableCompression: true

- S3 data is already compressed
- Saves CPU cycles

Full Example with Worker Pool:

package main

import (
"context"
"fmt"
"sync"

      "github.com/aws/aws-sdk-go-v2/service/s3"

)

type CopyJob struct {
SourceBucket string
SourceKey string
DestBucket string
DestKey string
}

func ProcessWithWorkerPool(ctx context.Context, jobs []CopyJob, numWorkers int) error {
// Create ONE shared S3 client (reuses connections)
s3Client, err := NewOptimizedS3Client(ctx)
if err != nil {
return err
}

      // Create job channel
      jobChan := make(chan CopyJob, numWorkers*2) // Buffer for efficiency

      // WaitGroup for workers
      var wg sync.WaitGroup

      // Start workers - all sharing the same S3 client
      for i := 0; i < numWorkers; i++ {
          wg.Add(1)
          go func(workerID int) {
              defer wg.Done()

              for job := range jobChan {
                  err := copyObject(ctx, s3Client, job)
                  if err != nil {
                      fmt.Printf("Worker %d: Error copying %s: %v\n",
                          workerID, job.SourceKey, err)
                      // TODO: Add to retry queue
                  }
              }
          }(i)
      }

      // Feed jobs to workers
      go func() {
          for _, job := range jobs {
              jobChan <- job
          }
          close(jobChan)
      }()

      // Wait for all workers to finish
      wg.Wait()

      return nil

}

func copyObject(ctx context.Context, client \*s3.Client, job CopyJob) error {
copySource := fmt.Sprintf("%s/%s", job.SourceBucket, job.SourceKey)

      _, err := client.CopyObject(ctx, &s3.CopyObjectInput{
          CopySource: &copySource,
          Bucket:     &job.DestBucket,
          Key:        &job.DestKey,
      })

      return err

}

Why This Matters:

Without connection pooling (default settings):

- Each request: TCP handshake (50ms) + TLS handshake (100ms) + request (50ms) = 200ms
- 780M objects = 4.9 years! 😱

With connection pooling (optimized):

- Reused connection: 0ms + 0ms + request (50ms) = 50ms
- 780M objects = 1.2 years (still long, but 4x faster!)

With connection pooling + 100 workers:

- Parallel execution: 50ms / 100 = 0.5ms effective
- 780M objects = ~4.5 days ✅

Additional Tips:

1. Use a single S3 client instance

// GOOD - One client shared by all workers
s3Client := NewOptimizedS3Client(ctx)
for i := 0; i < 100; i++ {
go worker(s3Client) // All workers share the client
}

// BAD - Creating new client per worker
for i := 0; i < 100; i++ {
client := NewOptimizedS3Client(ctx) // Wastes connections!
go worker(client)
}

2. Monitor connection reuse

// Add to your Transport config
Transport: &http.Transport{
// ... other settings ...

      // Log connection creation (for debugging)
      DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
          fmt.Printf("Creating new connection to %s\n", addr)
          return (&net.Dialer{
              Timeout:   30 * time.Second,
              KeepAlive: 30 * time.Second,
          }).DialContext(ctx, network, addr)
      },

}

You should see "Creating new connection" only ~100-200 times at startup, then nothing. If you see it continuously, connection pooling isn't working.

3. Use S3 Transfer Acceleration endpoint (if needed)

For even better performance across regions, but you're same-region so not needed.
