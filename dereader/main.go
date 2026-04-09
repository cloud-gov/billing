// Package dereader reads AWS Data Exports, specifically of the CUR 2.0 variety for now
package main

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

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

type gatherer struct {
	ctx context.Context

	exp     *btypes.Export
	expArn  *string
	dest    *btypes.S3Destination
	grain   string
	query   *string
	updated *time.Time

	period time.Time
	prefix string

	be *bcmdataexports.Client
	s3 *s3.Client
	tm *tm.Client
}

func (g *gatherer) FilterObject(o stypes.Object) bool {
	post := strings.ReplaceAll(*o.Key, g.prefix, "")
	path := strings.Contains(post, "/")
	return !path
}

func (g *gatherer) setPrefix() {
	g.prefix = fmt.Sprintf("%v/data/billing_period=%v/", *g.exp.Name, g.period.Format("2006-01"))
	if g.dest.S3Prefix != nil {
		g.prefix = fmt.Sprintf("%v/%v", *g.dest.S3Prefix, g.prefix)
	}
}

func (g *gatherer) getExport() (*bcmdataexports.GetExportOutput, error) {
	o, e := g.be.GetExport(g.ctx, &bcmdataexports.GetExportInput{ExportArn: g.expArn})
	if e != nil {
		return nil, e
	}

	g.exp = o.Export
	g.dest = o.Export.DestinationConfigurations.S3Destination
	g.query = o.Export.DataQuery.QueryStatement
	g.updated = o.ExportStatus.LastRefreshedAt
	g.grain = slices.Collect(maps.Values(
		o.Export.DataQuery.TableConfigurations))[0]["TIME_GRANULARITY"]

	g.setPrefix()

	return o, nil
}

func (g *gatherer) getFiles(path string) (*tm.DownloadDirectoryOutput, error) {
	return g.tm.DownloadDirectory(g.ctx, &tm.DownloadDirectoryInput{
		Bucket:      g.dest.S3Bucket,
		KeyPrefix:   &g.prefix,
		Destination: &path,
		Filter:      g,
	})
}

func newGatherer(ctx context.Context, cfg aws.Config, period time.Time) *gatherer {
	s3c := s3.NewFromConfig(cfg)
	g := gatherer{
		s3: s3c,
		tm: tm.New(s3c),
		be: bcmdataexports.NewFromConfig(cfg),

		ctx:    ctx,
		period: period,
		expArn: new(exportARN),
	}
	return &g
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

	gat := newGatherer(ctx, cfg, period)

	_, err = gat.getExport()
	check(err, errors.New("getting export"))

	out, err := gat.getFiles(".tmp")
	check(err, errors.New("downloading directory"))

	fmt.Println(out)
}
