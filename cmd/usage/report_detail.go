package main

import "fmt"

type (
	Report struct {
		UCreditSum int
		ReportLink
		ReportWriter
	}
	ReportNode struct {
		Path       string
		Slug       string
		Kind       string
		UCreditSum int
		Meters     []*ReportLeaf
		ReportLink
	}
	ReportLeaf struct {
		Kind       string
		UCreditUse int
		ReportLink
	}
	ReportLink struct {
		parent ReportLinker
		root   *Report
		Nodes  []*ReportNode
		ReportLinker
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

	if kk, ok := kind.(Kind); ok && kk.isMeter() { // this could also be implicit by excluding a name & path?
		rp = &ReportLeaf{Kind: kk.meterName(), UCreditUse: uCredits}
	} else if ok {
		rp = &ReportNode{Kind: kk.String(), UCreditSum: uCredits, Slug: name, Path: path}
	} else if ks, ok := kind.(string); ok {
		rp = &ReportNode{Kind: ks, UCreditSum: uCredits, Slug: name, Path: path}
	} else {
		return nil, fmt.Errorf("Report SetNode: kind must be stringable, got: %T", kind)
	}

	rl := rp.(ReportLinker)
	link.addChild(rl)

	rp.setRoot(r)
	rp.setParent(link)

	return rl, nil
}

func NewReporter() *Report {
	report := &Report{}
	report.root = report
	return report
}
