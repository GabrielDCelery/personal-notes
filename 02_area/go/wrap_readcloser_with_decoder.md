```go
package s3

import (
	"compress/flate"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"strings"
)

type decodedReadCloser struct {
	decoder io.ReadCloser
	body    io.ReadCloser
}

func wrapWithDecoder(body io.ReadCloser, encoding string) (io.ReadCloser, error) {
	switch strings.ToLower(strings.TrimSpace(encoding)) {
	case "gzip":
		gz, err := gzip.NewReader(body)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		return &decodedReadCloser{decoder: gz, body: body}, nil
	case "deflate":
		fl := flate.NewReader(body)
		return &decodedReadCloser{decoder: fl, body: body}, nil
	case "":
		// no encoding so just returning body because it is also implementing ReadCloser
		return body, nil
	default:
		return nil, fmt.Errorf("unsupported encoding: %s", encoding)
	}
}

func (d *decodedReadCloser) Read(p []byte) (int, error) {
	return d.decoder.Read(p)
}

func (d *decodedReadCloser) Close() error {
	return errors.Join(d.decoder.Close(),d.body.Close())
	var errs []error
	// if we have a decoder close the decoder
	if c, ok := d.decoder.(io.Closer); ok {
		if err := c.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	// close the body
	if err := d.body.Close(); err != nil {
		errs = append(errs, err)
	}
	//return errors if any
	return errors.Join(errs...)
}
```
