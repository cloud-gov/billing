// Package dereader reads AWS Data Exports, specifically of the CUR 2.0 variety for now
package main

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/gocarina/gocsv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	tm "github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"

	"github.com/aws/aws-sdk-go-v2/service/bcmdataexports"
	btypes "github.com/aws/aws-sdk-go-v2/service/bcmdataexports/types"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	stypes "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

var (
	ErrFileOpen  = errors.New("could not open file")
	ErrFileClose = errors.New("could not close file")
	ErrReading   = errors.New("problem reading CSV")
	ErrAWSConfig = errors.New("configuring AWS")
)

const exportARN string = "arn:aws:bcm-data-exports:us-east-1:223822404828:export/cg-billing-focus12-hourly-csv-4ecb05a5-f3c8-46ad-a097-1910cd50673a"

func check(e ...any) {
	if e[0] == nil {
		return
	}
	format := strings.Join(slices.Repeat([]string{"%w"}, len(e)), ": ")
	slices.Reverse(e)
	panic(fmt.Errorf(format, e...))
}

type checker func(e error, s ...string)

func (c checker) withLabels(l ...string) checker {
	return func(e error, s ...string) {
		c(e, append(l, s...)...)
	}
}

func getChecker(l ...string) checker {
	return func(e error, s ...string) {
		msg := strings.Join(append(l, s...), ": ")
		check(e, errors.New(msg))
	}
}

// BeeKeeper is a tool for getting and consuming AWS
// Billing & Cost Management (Bcm) Export Executions
type BeeKeeper struct {
	ctx context.Context

	exp     *btypes.Export
	expArn  *string
	dest    *btypes.S3Destination
	grain   string
	query   *string
	updated *time.Time

	period time.Time
	prefix string

	localPath string

	be *bcmdataexports.Client
	s3 *s3.Client
	tm *tm.Client

	check checker
}

func NewBeeKeeper(ctx context.Context, cfg aws.Config, path string, period time.Time) *BeeKeeper {
	s3c := s3.NewFromConfig(cfg)
	b := BeeKeeper{
		s3: s3c,
		tm: tm.New(s3c),
		be: bcmdataexports.NewFromConfig(cfg),

		ctx:       ctx,
		period:    period,
		localPath: path,

		expArn: new(exportARN),
		check:  getChecker("BeeKeeper"),
	}
	return &b
}

func (b *BeeKeeper) FilterObject(o stypes.Object) bool {
	post := strings.ReplaceAll(*o.Key, b.prefix, "")
	path := strings.Contains(post, "/")
	return !path
}

func (b *BeeKeeper) setPrefix() {
	b.prefix = fmt.Sprintf("%v/data/billing_period=%v/", *b.exp.Name, b.period.Format("2006-01"))
	if b.dest.S3Prefix != nil {
		b.prefix = fmt.Sprintf("%v/%v", *b.dest.S3Prefix, b.prefix)
	}
}

func (b *BeeKeeper) getExport() (*bcmdataexports.GetExportOutput, error) {
	o, e := b.be.GetExport(b.ctx, &bcmdataexports.GetExportInput{ExportArn: b.expArn})
	b.check(e)

	b.exp = o.Export
	b.dest = o.Export.DestinationConfigurations.S3Destination
	b.query = o.Export.DataQuery.QueryStatement
	b.updated = o.ExportStatus.LastRefreshedAt
	b.grain = slices.Collect(maps.Values(
		o.Export.DataQuery.TableConfigurations))[0]["TIME_GRANULARITY"]

	b.setPrefix()

	return o, nil
}

func (b *BeeKeeper) getFiles() (*tm.DownloadDirectoryOutput, error) {
	o, e := b.tm.DownloadDirectory(b.ctx, &tm.DownloadDirectoryInput{
		Bucket:      b.dest.S3Bucket,
		KeyPrefix:   &b.prefix,
		Destination: &b.localPath,
		Filter:      b,
	})
	b.check(e)
	return o, e
}

func (b *BeeKeeper) readLinesChan(lc chan *FocusSpec) {
	for ln := range lc {
		fmt.Println(ln.ConsumedQuantity.Decimal.String(), ln.ConsumedUnit)
	}
}

func (b *BeeKeeper) readFiles() error {
	lc := make(chan *FocusSpec)
	fsys := os.DirFS(b.localPath)
	chkErr := b.check.withLabels("readFiles")

	err := fs.WalkDir(fsys, ".", func(path string, dir fs.DirEntry, err error) error {
		chkErr(err, "walking dir, inner")

		if path == "." {
			return nil
		}

		file, err := fsys.Open(path)
		chkErr(err, fmt.Sprintf("could not open: '%s'", path))

		zr, err := gzip.NewReader(file)
		chkErr(err, fmt.Sprintf("could not unzip: '%s'", path))

		go b.readLinesChan(lc)

		err = gocsv.UnmarshalToChan(zr, lc)
		chkErr(err, "scanning")

		return nil
	})
	chkErr(err, "walking dir, outer")

	return nil
}

func main() {
	defer (func() {
		if r := recover(); r != nil {
			fmt.Printf("dereader: %v", r)
		}
	})()

	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx)
	check(err, ErrAWSConfig)

	// TODO: get date of last aws sync, trunc month, use as this date prefix
	// could also potentially use checksum to detect change
	period := time.Now()
	filePath := ".tmp"

	bkeep := NewBeeKeeper(ctx, cfg, filePath, period)

	_, err = bkeep.getExport()
	check(err, errors.New("getting export"))

	out, err := bkeep.getFiles(".tmp")
	check(err, errors.New("downloading directory"))

	err = bkeep.readFiles()
	check(err, errors.New("reading BEE results"))
}
