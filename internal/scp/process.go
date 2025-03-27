package scp

import (
    "fmt"
    "github.com/schollz/progressbar/v3"
    "io"
)

type ProgressReader struct {
    reader     io.Reader
    bar        *progressbar.ProgressBar
    totalBytes int64
}

func NewProgressReader(reader io.Reader, size int64, description string) *ProgressReader {
    bar := progressbar.NewOptions64(
        size,
        progressbar.OptionSetDescription(description),
        progressbar.OptionShowBytes(true),
        progressbar.OptionSetWidth(30),
        progressbar.OptionShowCount(),
        progressbar.OptionOnCompletion(func() {
            fmt.Printf("\n")
        }),
    )

    return &ProgressReader{
        reader:     reader,
        bar:        bar,
        totalBytes: size,
    }
}

func (pr *ProgressReader) Read(p []byte) (int, error) {
    n, err := pr.reader.Read(p)
    if n > 0 {
        pr.bar.Add(n)
    }
    return n, err
}

type ProgressWriter struct {
    writer     io.Writer
    bar        *progressbar.ProgressBar
    totalBytes int64
}

func NewProgressWriter(writer io.Writer, size int64, description string) *ProgressWriter {
    bar := progressbar.NewOptions64(
        size,
        progressbar.OptionSetDescription(description),
        progressbar.OptionShowBytes(true),
        progressbar.OptionSetWidth(30),
        progressbar.OptionShowCount(),
        progressbar.OptionOnCompletion(func() {
            fmt.Printf("\n")
        }),
    )

    return &ProgressWriter{
        writer:     writer,
        bar:        bar,
        totalBytes: size,
    }
}

func (pw *ProgressWriter) Write(p []byte) (int, error) {
    n, err := pw.writer.Write(p)
    if n > 0 {
        pw.bar.Add(n)
    }
    return n, err
}