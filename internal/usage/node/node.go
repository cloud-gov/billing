// Package node provides functionality to support creeating ResourceNodes
package node

import (
	"errors"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
)

type Node struct {
	Slug string
	Path string

	CustomerID        pgtype.UUID
	ResourceNaturalID string // e.g. an CF App ID, a Workshop namespace ID; may relate to multiple Resources
}

type NodeOpt func(*Node) error

var (
	ErrSlugInvalid  error = errors.New("slug is invalid, cannot contain `.`")
	ErrPathSlugless error = errors.New("cannot create path without slug")
)

var (
	saniSlugExpr *regexp.Regexp = regexp.MustCompile(`[^\w]`)
	saniPathExpr *regexp.Regexp = regexp.MustCompile(`[^\w.]`)
)

// WithSlugAuto combines slugParts with `_`, re-parts on `.`, and strips [^\w]
func WithSlugAuto(slugParts ...string) NodeOpt {
	return func(n *Node) error {
		b := strings.Builder{}
		for i, s := range slugParts {
			if i > 0 {
				b.WriteString("_")
			}
			s = strings.ReplaceAll(s, ".", "_")
			s = saniSlugExpr.ReplaceAllString(s, "_")
			b.WriteString(s)
		}
		n.Slug = b.String()
		return nil
	}
}

// WithPathAuto combines pathParts with `.`, strips [^\w.], and appends slug
func WithPathAuto(pathParts ...string) NodeOpt {
	return func(n *Node) error {
		b := strings.Builder{}
		for i, s := range pathParts {
			if i > 0 {
				b.WriteString(".")
			}
			b.WriteString(saniPathExpr.ReplaceAllString(s, "_"))
		}
		if n.Slug == "" {
			return ErrPathSlugless
		}
		b.WriteString("." + n.Slug)
		n.Path = b.String()
		return nil
	}
}

// WithPathByParent calls WithPathAuth with a parent node
func WithPathByParent(parent *Node) NodeOpt {
	return WithPathAuto(parent.Path)
}

func New(customerID any, resourceID string, opts ...NodeOpt) (*Node, error) {
	n := &Node{ResourceNaturalID: resourceID}

	if customerID, ok := customerID.(pgtype.UUID); !ok {
		u := pgtype.UUID{}
		if err := u.Scan(customerID); err != nil {
			return nil, err
		}
		n.CustomerID = u
	} else {
		n.CustomerID = customerID
	}

	if !n.CustomerID.Valid {
		err := n.CustomerID.Scan("FFFFFFFF-FFFF-FFFF-FFFF-FFFFFFFFFFFF")
		if err != nil {
			panic(err)
		}
	}

	for _, o := range opts {
		if err := o(n); err != nil {
			return nil, err
		}
	}
	return n, nil
}
