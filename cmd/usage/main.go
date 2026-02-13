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
	"github.com/jackc/pgx/v5/pgtype"
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
	flag.StringVar(&cid, "cid", "", "Narrow scope to Customer by ID, falls back to $CG_USAGE_CUSTOMER_ID if neither -cid or -cname defined")
	flag.StringVar(&cname, "cname", "", "Narrow results to Customer by name")
	flag.StringVar(&lquery, "lq", "", "Provide an `lquery` to search with; supercedes org & space")
	flag.StringVar(&org, "org", "", "Filter by org name")
	flag.StringVar(&space, "space", "", "Filter by space same")
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
	n, err := getNodes(ctx, q, nodeQuery, customerID)
	if err != nil {
		return fmtErr(ErrGettingNodes, err)
	}

	logger.Debug("run: got usage", "usage", n)

	return err
}

func getNodes(ctx context.Context, q db.Querier, query string, customerID pgtype.UUID) ([]db.GetUsageByPathRow, error) {
	return q.GetUsageByPath(ctx, db.GetUsageByPathParams{
		Path:       query,
		CustomerID: customerID,
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

	if cs, e := q.GetCustomersByName(ctx, cname); e != nil {
		err = fmtErr(ErrGetCustomer, e)
	} else if len(cs) > 1 {
		err = fmt.Errorf("got more than one customer for '%v'", cname)
	} else {
		id = cs[0].ID
	}

	return id, err
}

func getRawCID() string {
	if cid == "" && cname == "" { // only fallback to env no user input
		return os.Getenv("CG_USAGE_CUSTOMER_ID")
	}
	return cid
}
