package main

import "fmt"

type (
	Report struct {
		UCreditSum int `json:"microcredit_sum"`
		ReportLink
		ReportWriter
	}
	ReportNode struct {
		UCreditSum int           `json:"microcredit_sum"`
		Meters     []*ReportLeaf `json:",omitempty"`
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

		Slug  string
		Path  string
		Kind  string
		Nodes []*ReportNode `json:",omitempty"`
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

	var isLeaf bool
	var stKind string

	switch k := kind.(type) {
	case Kind:
		isLeaf = k.isMeter()
		stKind = k.String()
	case string:
		stKind = k
	default:
		return nil, fmt.Errorf("Report SetNode: kind must be stringable, got: %T", kind)
	}

	rl := ReportLink{Kind: stKind, Slug: name, Path: path}

	if isLeaf {
		rp = &ReportLeaf{UCreditUse: uCredits, ReportLink: rl}
	} else {
		rp = &ReportNode{UCreditSum: uCredits, ReportLink: rl}
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
