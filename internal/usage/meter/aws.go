package meter

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/cloudfoundry/go-cfclient/v3/resource"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	"github.com/aws/smithy-go/logging"

	"github.com/cloud-gov/billing/internal/db"

	"github.com/cloud-gov/billing/internal/usage/node"
	"github.com/cloud-gov/billing/internal/usage/reader"
)

type AWSMeterDB interface {
	GetCFOrg(ctx context.Context, id pgtype.UUID) (db.CFOrg, error)
}

// AWSMeter reads usage from AWS
type AWSMeter struct {
	logger *slog.Logger
	client AWSMeterCfProvider
	dbq    AWSMeterDB
}

func NewAWSMeter(
	logger *slog.Logger, client AWSMeterCfProvider, dbq ServiceMeterDB,
) *AWSMeter {
	return &AWSMeter{
		logger: logger.WithGroup("AWSMeter"),
		client: client,
		dbq:    dbq,
	}
}

func (m *AWSMeter) Name() string {
	return "aws"
}

// AWSFinding is a temporary stand-in for scaffolding
// TODO: delete this!
type AWSFinding struct {
	someGUID string
}

// smogger returns an adaptor for AWS smithy logging.Logger to SLog
func (m *AWSMeter) smogger(ctx context.Context) logging.Logger {
	return logging.LoggerFunc(func(xlevel logging.Classification, format string, v ...any) {
		var slevel slog.Level

		switch xlevel {
		case logging.Debug:
			slevel = slog.LevelDebug
		case logging.Warn:
			slevel = slog.LevelWarn
		}

		m.logger.Log(ctx, slevel, fmt.Sprintf(format, v...))
	})
}

func getTimePeriodDays(end, diff int) *types.DateInterval {
	now := time.Now().UTC()
	periodFmt := "2006-01-02"

	nTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	eTime := nTime.AddDate(0, 0, end)
	sTime := nTime.AddDate(0, 0, diff)

	eStr := eTime.Format(periodFmt)
	sStr := sTime.Format(periodFmt)

	return &types.DateInterval{
		End:   &eStr,
		Start: &sStr,
	}
}

// ReadUsage returns the point-in-time usage of AWS services not metered through Cloud Foundry
// Returns a non-nil error if there was an error during the overall process of reading usage information from the target system. If individual readings had errors, their errs fields should be set.
func (m *AWSMeter) ReadUsage(ctx context.Context) ([]*reader.Measurement, []*node.Node, error) {
	m.logger.DebugContext(ctx, "configuring aws")
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithLogger(m.smogger(ctx)),
		config.WithUseFIPSEndpoint(aws.FIPSEndpointStateDisabled),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("aws meter loading aws config: %w", err)
	}

	m.logger.DebugContext(ctx, "getting cost & usage")
	ce := costexplorer.NewFromConfig(cfg)
	out, err := getCostAndUsage(ctx, ce)
	if err != nil {
		return nil, nil, fmt.Errorf("getting cost & usage: %w", err)
	}

	fmt.Printf("a: %#v\n", out.DimensionValueAttributes)
	fmt.Printf("b: %#v\n", out.GroupDefinitions)
	fmt.Printf("c: %#v\n", out.ResultMetadata)
	fmt.Printf("d: %#v\n", out.NextPageToken)
	fmt.Printf("e: %#v\n", out.ResultsByTime)

	m.logger.DebugContext(ctx, "got cost data!", "out", out)

	findings := make([]AWSFinding, 5)

	usage := make([]*reader.Measurement, len(findings))
	nodes := make([]*node.Node, 0, len(findings))

	// For caching lookups, key is GUID
	lookups := struct {
		customers    map[string]*db.Customer
		orgs         map[string]*resource.Organization
		spaces       map[string]*resource.Space
		svcs         map[string]*resource.ServiceInstance
		svcPlans     map[string]*resource.ServicePlan
		svcOfferings map[string]*resource.ServiceOffering
	}{}

	for i, instance := range findings {
		// TODO:
		// - Get org, space, service, plan & offering as necessary
		// - Should this meter create resource nodes? if not, need to handle outside

		// TODO: this is just scaffolding, these lookups are nonsense
		usage[i] = &reader.Measurement{
			Meter:                 m.Name(),
			CustomerID:            lookups.customers[instance.someGUID].ID,
			OrgID:                 lookups.orgs[instance.someGUID].GUID,
			OrgName:               lookups.orgs[instance.someGUID].Name,
			ResourceKindNaturalID: lookups.svcPlans[instance.someGUID].GUID,
			ResourceNaturalID:     lookups.svcs[instance.someGUID].GUID,
			Value:                 1,
			Errs:                  nil,
		}
	}

	return usage, nodes, nil
}

func getCostAndUsage(ctx context.Context, ce *costexplorer.Client) (*costexplorer.GetCostAndUsageOutput, error) {
	return ce.GetCostAndUsage(ctx, &costexplorer.GetCostAndUsageInput{
		Granularity: types.GranularityDaily,
		GroupBy: []types.GroupDefinition{{
			Key:  new("Instance GUID"),
			Type: types.GroupDefinitionTypeTag,
		}},
		Metrics: []string{
			"AmortizedCost", "BlendedCost", "NetAmortizedCost",
			"NetUnblendedCost", "NormalizedUsageAmount", "UnblendedCost", "UsageQuantity",
		},
		Filter: &types.Expression{
			Dimensions: &types.DimensionValues{
				Key: types.DimensionUsageTypeGroup,
				Values: []string{
					"RDS: Storage",
					"S3: Storage - Standard", // only standard used by customers right now
				},
				MatchOptions: []types.MatchOption{types.MatchOptionEquals, types.MatchOptionCaseSensitive},
			},
		},
		TimePeriod: getTimePeriodDays(0, -1),
	})
}
