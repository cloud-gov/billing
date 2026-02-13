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
	"strings"

	"github.com/cloud-gov/billing/internal/config"
	"github.com/cloud-gov/billing/internal/db"
	"github.com/cloud-gov/billing/internal/dbx"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrArgsShort           = errors.New("not enough args")
	ErrArgsNoCustomer      = errors.New("no customer ID supplied")
	ErrGetCustomer         = errors.New("getting customer")
	ErrArgsNoQuery         = errors.New("no query supplied")
	ErrBadConfig           = errors.New("reading config from environment")
	ErrCFClient            = errors.New("creating Cloud Foundry client")
	ErrCFConfig            = errors.New("parsing Cloud Foundry connection configuration")
	ErrCrontab             = errors.New("parsing crontab for periodic job execution")
	ErrDBConn              = errors.New("connecting to database")
	ErrGettingNodes        = errors.New("getting nodes")
	ErrGettingMeasurements = errors.New("getting nodes")
)

func fmtErr(outer, inner error) error {
	return fmt.Errorf("%w: %w", outer, inner)
}

var (
	cid    string
	cname  string
	lquery string
	org    string
	space  string
)

func init() {
	flag.StringVar(&cid, "cid", "$CG_USAGE_CUSTOMER_ID", "Narrow scope to Customer by ID")
	flag.StringVar(&cname, "cname", "", "Narrow results to Customer by name")
	flag.StringVar(&lquery, "lq", "", "Provide an `lquery` to search with; supercedes org & space")
	flag.StringVar(&org, "org", "", "Filter by org name")
	flag.StringVar(&space, "space", "", "Filter by space same")
	flag.Parse()
}

func getNodes(ctx context.Context, q db.Querier, query, customerID string) ([]db.GetUsageByPathRow, error) {
	return q.GetUsageByPath(ctx, db.GetUsageByPathParams{
		Path:       query,
		CustomerID: dbx.UtilUUID(customerID),
	})
}

// func getMeasures(ctx context.Context, q db.Querier, nodes []db.ResourceNode) ([]db.Measurement, error) {
// }

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

	var customerID string
	if cid == "" && cname == "" {
		customerID = os.Getenv("CG_USAGE_CUSTOMER_ID")
	} else if cid != "" {
		customerID = cid
	} else if cname != "" {
		cs, err := q.GetCustomersByName(ctx, cname)
		if err != nil {
			return fmtErr(ErrGetCustomer, err)
		}
		if len(cs) > 1 {
			return fmt.Errorf("got more than one customer for '%v'", cname)
		}
		customerID = cs[0].ID.String()
	} else {
		panic("uh oh… no customer")
	}

	nodeQuery := strings.Builder{}
	nodeQuery.WriteString("apps.usage")
	if lquery != "" {
		nodeQuery.WriteString(lquery)
	} else {
		if org != "" {
			fmt.Fprintf(&nodeQuery, ".cforg_%v%%", org)
		} else {
			nodeQuery.WriteString(".cforg_%")
		}
		if space != "" {
			fmt.Fprintf(&nodeQuery, ".space_%v%%", space)
		} else {
			nodeQuery.WriteString(".space_%")
		}

		nodeQuery.WriteString(".*{1,}")
	}

	logger.Debug("run: getting usage", "customerID", customerID, "query", nodeQuery.String())
	n, err := getNodes(ctx, q, nodeQuery.String(), customerID)
	if err != nil {
		return fmtErr(ErrGettingNodes, err)
	}

	logger.Debug("run: got usage", "usage", n)

	return err
}

func main() {
	ctx := context.Background()
	out := os.Stdout
	err := run(ctx, out)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
