package main

import "fmt"

type (
	Report struct {
		UCreditSum int `json:"microcredit_sum"`
		ReportLink
		ReportWriter `json:"-"`
	}
	ReportNode struct {
		UCreditSum int           `json:"microcredit_sum"`
		Meters     []*ReportLeaf `json:"meters,omitempty"`
		ReportLink
	}
	ReportLeaf struct {
		UCreditUse int `json:"microcredit_use"`
		ReportLink
	}
	ReportLink struct {
		ReportLinker `json:"-"`

		root   *Report      `json:"-"`
		parent ReportLinker `json:"-"`

		Slug  string        `json:"slug,omitempty"`
		Path  string        `json:"path,omitempty"`
		Kind  string        `json:"kind,omitempty"`
		Nodes []*ReportNode `json:"nodes,omitempty"`
	}
)

func (n *ReportNode) getChildren() []ReportLinker {
	if len(n.Meters) > 0 {
		return sliceToReportLinkers(n.Meters)
	} else {
		return sliceToReportLinkers(n.Nodes)
	}
}

func (n *ReportNode) addChild(link ReportLinker) {
	switch l := link.(type) {
	case *ReportLeaf:
		n.Meters = append(n.Meters, l)
	case *ReportNode:
		n.Nodes = append(n.Nodes, l)
	}
}

func (rl *ReportLink) getParent() ReportLinker {
	if rl.parent != nil {
		return rl.parent
	} else {
		return rl.root
	}
}

func (rl *ReportLink) addChild(link ReportLinker) {
	rl.Nodes = append(rl.Nodes, link.(*ReportNode))
}

func (rl *ReportLink) getChildren() []ReportLinker { return sliceToReportLinkers(rl.Nodes) }
func (rl *ReportLink) setRoot(r *Report)           { rl.root = r }
func (rl *ReportLink) setParent(r ReportLinker)    { rl.parent = r }

func (r *Report) SetNode(link ReportLinker, uCredits int, kind any, name, path string) (ReportLinker, error) {
	var rp rootParenter
	var stKind string

	// TODO: meters/leaves are not currently attached
	// - These are really just the branch tips
	// - See cloud-gov/cg-interface/cg-billing#89
	switch k := kind.(type) {
	case Kind:
		stKind = k.String()
	case string:
		stKind = k
	default:
		return nil, fmt.Errorf("Report SetNode: kind must be stringable, got: %T", kind)
	}

	rp = &ReportNode{
		UCreditSum: uCredits,
		ReportLink: ReportLink{Kind: stKind, Slug: name, Path: path},
	}

	linker := rp.(ReportLinker)
	link.addChild(linker)

	rp.setRoot(r)
	rp.setParent(link)

	return linker, nil
}

func NewReporter() *Report {
	report := &Report{}
	report.root = report
	return report
}
