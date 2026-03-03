// Package main for logging out usage data
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"regexp"
	"slices"
	"strings"

	"github.com/cloud-gov/billing/internal/config"
	"github.com/cloud-gov/billing/internal/db"
	"github.com/cloud-gov/billing/internal/dbx"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrGetCustomer    = errors.New("getting customer")
	ErrBadConfig      = errors.New("reading config from environment")
	ErrDBConn         = errors.New("connecting to database")
	ErrGettingNodes   = errors.New("getting nodes")
	ErrCreatingReport = errors.New("making report")
)

func fmtErr(outer, inner error) error {
	return fmt.Errorf("%w: %w", outer, inner)
}

var (
	appReg = regexp.MustCompile(`^app_[^\.]+$`)
	svcReg = regexp.MustCompile(`^svc_[^\.]+$`)
)

func isApp(s pgtype.Text) bool {
	return appReg.MatchString(s.String)
}

func isService(s pgtype.Text) bool {
	return svcReg.MatchString(s.String)
}

var (
	cid    string
	cname  string
	lquery string
	org    string
	space  string
	after  int
	before int
	period string
)

func init() {
	flag.StringVar(&cid, "cid", "", "Narrow scope to Customer by ID, falls back to $CG_USAGE_CUSTOMER_ID if neither -cid or -cname defined")
	flag.StringVar(&cname, "cname", "", "Narrow results to Customer by name")
	flag.StringVar(&lquery, "lq", "", "Provide an `lquery` to search with; supercedes org & space")
	flag.StringVar(&org, "org", "", "Filter by org name")
	flag.StringVar(&space, "space", "", "Filter by space same")
	flag.IntVar(&after, "a", -1, "Filter [a]fter n-periods")
	flag.IntVar(&before, "b", 0, "Filter [b]efore n-periods")
	flag.StringVar(&period, "p", "month", "Time [p]eriod/interval, e.g. month, week, day")
	flag.Parse()
}

// func getMeasures(ctx context.Context, q db.Querier, nodes []db.ResourceNode) ([]db.Measurement, error) {
// }

func main() {
	ctx := context.Background()
	out := os.Stdout
	err := run(ctx, out)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func run(ctx context.Context, out io.Writer) error {
	c, err := config.New()
	if err != nil {
		return fmtErr(ErrBadConfig, err)
	}

	logger := slog.New(slog.NewJSONHandler(out, &slog.HandlerOptions{
		Level: c.LogLevel,
	}))

	logger.Debug("run: initializing database")
	conn, err := pgxpool.New(ctx, "") // Pass empty connString so PG* environment variables will be used.
	if err != nil {
		return fmtErr(ErrDBConn, err)
	}

	q := dbx.NewQuerier(db.New(conn))

	customerID, err := getCustomerID(ctx, q)
	if err != nil {
		return err
	}

	nodeQuery := buildQuery()

	logger.Debug("run: getting usage", "customerID", customerID, "query", nodeQuery)
	nodes, err := getNodes(ctx, q, nodeQuery, customerID)
	if err != nil {
		return fmtErr(ErrGettingNodes, err)
	}
	logger.Debug("run: got usage", "usage", nodes)

	logger.Debug("run: making report")
	report := NewReporter()
	var link ReportLinker

	slices.Reverse(nodes)
	for i, n := range nodes {
		uCreds, err := n.TotalMicrocredits.Int64Value()
		if err != nil {
			return fmtErr(ErrCreatingReport, err)
		}
		uCredsInt := int(uCreds.Int64)

		if i == 0 {
			report.UCreditSum = uCredsInt
			continue
		}

		p := nodes[i-1]

		paths := []string{}
		levels := []pgtype.Text{n.L1, n.L2, n.L3, n.L4}
		for _, l := range levels {
			if !l.Valid {
				break
			}
			paths = append(paths, l.String)
		}
		path := strings.Join(paths, ".")

		// org
		if n.L1.Valid && !n.L2.Valid {
			// org is always linked to root
			link, err = report.SetNode(report, uCredsInt, Org, n.L1.String, path)
			if err != nil {
				return fmtErr(ErrCreatingReport, err)
			}
			continue
		}

		// space generalized
		if n.L2.Valid && !n.L3.Valid {
			if p.L4.Valid { // go back from leaf > space/s > space/g > org
				link = link.getParent().getParent().getParent()
			} else if p.L3.Valid { // go back from space/s > space/g > org
				link = link.getParent().getParent()
			} else if p.L2.Valid { // go space/g > org
				link = link.getParent()
			}
			link, err = report.SetNode(link, uCredsInt, Space, n.L2.String, path)
			if err != nil {
				return fmtErr(ErrCreatingReport, err)
			}
			continue
		}

		// space specific
		if n.L3.Valid && !n.L4.Valid {
			if p.L4.Valid { // go back from leaf > space/s > space/g
				link = link.getParent().getParent()
			} else if p.L3.Valid { // go space/s > space/g
				link = link.getParent()
			}
			link, err = report.SetNode(link, uCredsInt, Space, n.L3.String, path)
			if err != nil {
				return fmtErr(ErrCreatingReport, err)
			}
			continue
		}

		if n.L4.Valid {
			if p.L4.Valid { // go from leaf > space/s
				link = link.getParent()
			}

			var k Kind
			if isApp(n.L4) {
				k = CfApp
			} else if isService(n.L4) {
				k = CfSvc
			}

			link, err = report.SetNode(link, uCredsInt, k, n.L4.String, path)
			if err != nil {
				return fmtErr(ErrCreatingReport, err)
			}
			continue
		}

		logger.Debug("weirds gotten in report", "node", n)
	}
	logger.Debug("run: got report", "report", report)

	return err
}

func getNodes(ctx context.Context, q db.Querier, query string, customerID pgtype.UUID) ([]db.GetUsageByPathRow, error) {
	return q.GetUsageByPath(ctx, db.GetUsageByPathParams{
		Path:       query,
		CustomerID: customerID,
		Before:     int32(before),
		After:      int32(after),
		Period:     period,
	})
}

func buildQuery() string {
	nodeQuery := strings.Builder{}
	nodeQuery.WriteString("apps.usage")

	if lquery != "" {
		nodeQuery.WriteString(lquery)
		return nodeQuery.String()
	}

	// L1
	if org != "" {
		fmt.Fprintf(&nodeQuery, ".cforg_%v%%", org)
	} else {
		nodeQuery.WriteString(".cforg_%")
	}

	// L2/3
	if space != "" {
		fmt.Fprintf(&nodeQuery, ".space_%v%%", space)
	} else {
		nodeQuery.WriteString(".space_%")
	}

	// Resources Leaves
	nodeQuery.WriteString(".*{1,}")

	return nodeQuery.String()
}

func getCustomerID(ctx context.Context, q dbx.Querier) (id pgtype.UUID, err error) {
	r := getRawCID()
	if r != "" {
		return dbx.UtilUUID(r), nil
	}

	cs, e := q.GetCustomersByName(ctx, cname)
	if e != nil {
		return id, fmtErr(ErrGetCustomer, e)
	} else if len(cs) == 1 {
		return cs[0].ID, err
	} else if len(cs) > 1 {
		cList := &strings.Builder{}
		fmt.Fprint(cList, "found customers:\nname\tuuid\n")
		for _, c := range cs {
			fmt.Fprintf(cList, "%v\t%v\n", c.Name, c.ID)
		}
		slog.Error(cList.String())
	}

	return id, fmt.Errorf("found %v customers for '%v', need 1", len(cs), cname)
}

func getRawCID() string {
	if cid == "" && cname == "" { // only fallback to env no user input
		return os.Getenv("CG_USAGE_CUSTOMER_ID")
	}
	return cid
}
