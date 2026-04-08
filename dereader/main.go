// Package dereader reads AWS Data Exports, specifically of the CUR 2.0 variety for now
package main

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bcmdataexports"
	"github.com/aws/aws-sdk-go-v2/service/bcmdataexports/types"
)

var (
	ErrFileOpen  = errors.New("could not open file")
	ErrFileClose = errors.New("could not close file")
	ErrReading   = errors.New("problem reading CSV")
	ErrAWSConfig = errors.New("configuring AWS")
)

const exportARN string = "arn:aws:bcm-data-exports:us-east-1:223822404828:export/cg-billing-export-cur-hourly-csv-4d785f67-dee8-414d-b592-46c0a2c832dd"

func check(e ...any) {
	if e[0] == nil {
		return
	}
	format := strings.Join(slices.Repeat([]string{"%w"}, len(e)), ": ")
	slices.Reverse(e)
	panic(fmt.Errorf(format, e...))
}

func findMostRecentSuccessfulExecution(
	ctx context.Context,
	dec *bcmdataexports.Client,
	exp *types.Export,
	token *string,
) (*bcmdataexports.GetExecutionOutput, error) {
	execs, err := dec.ListExecutions(ctx, &bcmdataexports.ListExecutionsInput{
		ExportArn: exp.ExportArn,
		NextToken: token,
	})
	check(err, errors.New("listing export executions"))

	for _, e := range execs.Executions {
		if e.ExecutionStatus.StatusCode == types.ExecutionStatusCodeDeliverySuccess {
			exec, err := dec.GetExecution(ctx, &bcmdataexports.GetExecutionInput{
				ExportArn:   exp.ExportArn,
				ExecutionId: e.ExecutionId,
			})
			check(err, errors.New("getting full execution"))
			return exec, nil
		}
	}

	if execs.NextToken == nil {
		return nil, errors.New("no successful execution")
	}

	return findMostRecentSuccessfulExecution(ctx, dec, exp, execs.NextToken)
}

func main() {
	defer (func() {
		if r := recover(); r != nil {
			fmt.Printf("dereader: %v", r)
		}
	})()

	ctx := context.Background()

	// ReadCSV()

	cfg, err := config.LoadDefaultConfig(ctx)
	check(err, ErrAWSConfig)

	dec := bcmdataexports.NewFromConfig(cfg)

	exp, err := dec.GetExport(ctx, &bcmdataexports.GetExportInput{ExportArn: new(exportARN)})
	check(err, errors.New("getting export"))

	execs, err := dec.ListExecutions(ctx, &bcmdataexports.ListExecutionsInput{ExportArn: exp.Export.ExportArn})
	check(err, errors.New("getting export executions"))

	for _, e := range execs.Executions {
		if e.ExecutionStatus.StatusCode == types.ExecutionStatusCodeDeliverySuccess {
			// exec, err := dec.GetExecution(ctx, &bcmdataexports.GetExecutionInput{ExecutionId: e.ExecutionId})
			// Find reading, compare?
		}
	}
}
