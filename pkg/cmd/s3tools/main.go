package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/eirture/s3tools/pkg/config"
)

const (
	KeyDeleteAfterDays = "x-amz-extend-delete-after-days"
)

func NewS3(cred *config.Credential) *s3.S3 {
	cfg := &aws.Config{
		Region:   aws.String(cred.Region),
		Endpoint: aws.String(cred.Endpoint),
		Credentials: credentials.NewStaticCredentials(
			cred.AccessKeyID,
			cred.AccessKeySecret,
			cred.Token,
		),
		DisableSSL:                     aws.Bool(true), //如果endpoint没有指定http/https, 根据DisableSSL 字段来决定使用http or https
		S3ForcePathStyle:               aws.Bool(true),
		LogLevel:                       aws.LogLevel(aws.LogDebug),
		DisableRestProtocolURICleaning: aws.Bool(true), //可以保留 keyname 里的"//",  https://github.com/aws/aws-sdk-go/issues/2559
		S3Disable100Continue:           aws.Bool(true), // a bug see https://jira.qiniu.io/browse/KODO-8212
	}
	sess := session.Must(session.NewSessionWithOptions(session.Options{Config: *cfg}))
	return s3.New(sess)
}

func PubObject(ctx context.Context, cli *s3.S3, bucket, key string, src io.ReadSeeker, header map[string]string) (err error) {
	input := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   src,
	}
	out, err := cli.PutObjectWithContext(ctx, input, request.WithSetRequestHeaders(header))
	if err != nil {
		return
	}
	log.Println(out.String())
	return nil
}

func MultipartUpload(ctx context.Context, cli *s3.S3, bucket, key string, size, partSize int64, header map[string]string, c int) (err error) {
	mui := &s3.CreateMultipartUploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	muo, err := cli.CreateMultipartUpload(mui)
	if err != nil {
		return
	}
	uid := muo.UploadId

	// upload parts
	pn := size / partSize
	if size%partSize > 0 {
		pn += 1
	}

	if int64(c) > pn {
		c = int(pn)
	}

	parts := make([]*s3.CompletedPart, c)
	for i := int64(0); i < pn; i++ {
		pSize := partSize
		remain := (i + 1) * partSize
		if remain < pSize {
			pSize = remain
		}

		upi := &s3.UploadPartInput{
			Bucket:     aws.String(bucket),
			Key:        aws.String(key),
			PartNumber: aws.Int64(i),
			UploadId:   uid,
			Body:       NewFakeReadSeekCloser(pSize),
		}
		upo, upErr := cli.UploadPartWithContext(ctx, upi, request.WithSetRequestHeaders(header))
		if upErr != nil {
			return upErr
		}

		parts = append(parts, &s3.CompletedPart{
			ETag:       upo.ETag,
			PartNumber: upi.PartNumber,
		})
	}

	// complete upload parts
	cmuo, err := cli.CompleteMultipartUploadWithContext(ctx, &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(key),
		UploadId: uid,
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: parts,
		},
	}, request.WithSetRequestHeaders(header))
	if err != nil {
		return
	}
	log.Println(cmuo.String())

	return nil
}

type FakeReadSeekCloser struct {
	Size  int64
	count int64
}

func NewFakeReadSeekCloser(size int64) *FakeReadSeekCloser {
	return &FakeReadSeekCloser{
		Size: size,
	}
}

func (r *FakeReadSeekCloser) Read(p []byte) (n int, err error) {
	if r.count >= r.Size {
		return 0, io.EOF
	}

	n = len(p)
	if remain := r.Size - r.count; remain < int64(n) {
		n = int(remain)
	}
	for i := 0; i < n; i++ {
		p[i] = byte('Q')
	}
	r.count += int64(n)

	return
}

func (r *FakeReadSeekCloser) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		r.count = offset
	case io.SeekCurrent:
		r.count += offset
	case io.SeekEnd:
		r.count = r.Size + offset
	}

	return r.count, nil
}

func (r *FakeReadSeekCloser) Close() error {
	return nil
}

func makeReader(size int64) io.ReadSeekCloser {
	return NewFakeReadSeekCloser(size)
}

type Task struct {
	Bucket          string
	Key             string
	Size            int64
	DeleteAfterDays string
}

func (t *Task) String() string {
	return fmt.Sprintf("%s/%s %v", t.Bucket, t.Key, t.Size)
}

func upload(cfg *config.Config, cli *s3.S3, total int) (err error) {

	workerCount := int(cfg.Workers)
	if total < workerCount {
		workerCount = total
	}

	ch := make(chan *Task, workerCount)
	ctx := context.Background()

	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for t := range ch {
				rsc := makeReader(t.Size)
				log.Printf("upload: %s\n", t)
				pErr := PubObject(ctx, cli, t.Bucket, t.Key, rsc, map[string]string{
					KeyDeleteAfterDays: t.DeleteAfterDays,
				})
				if pErr != nil {
					log.Printf("upload error %s. err: %v\n", t, pErr)
					// continue
				}
			}
		}()
	}

	nts := strconv.FormatInt(time.Now().Unix(), 10)

	for i := 0; i < total; i++ {
		key := fmt.Sprintf("%s-%d", nts, i)
		size := cfg.FileSizes[rand.Intn(len(cfg.FileSizes))]
		ch <- &Task{
			Bucket:          cfg.Bucket,
			Key:             key,
			Size:            size,
			DeleteAfterDays: cfg.DeleteAfterDays[rand.Intn(len(cfg.DeleteAfterDays))],
		}
	}

	close(ch)
	wg.Wait()
	return
}

func main() {
	var (
		configPath = "s3tools.yaml"
		total      = 1
	)

	flag.StringVar(&configPath, "f", configPath, "the config path")
	flag.IntVar(&total, "n", total, "total files count")
	flag.Parse()

	rand.Seed(time.Now().UnixNano())

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatal(err)
	}

	if cfg.Credential == nil {
		log.Fatal("No credential found!")
	}
	cli := NewS3(cfg.Credential)

	err = upload(cfg, cli, total)
	if err != nil {
		log.Fatal(err)
	}
}
