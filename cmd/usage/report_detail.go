package usage

import "fmt"

type (
	Report struct {
		ReportLink
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

func (rl *ReportLink) getParent() ReportLinker {
	if rl.parent != nil {
		return rl.parent
	} else {
		return rl.root
	}
}

func (rl *ReportLink) getChildren() []ReportLinker { return sliceToReportLinkers(rl.Nodes) }
func (rl *ReportLink) setRoot(r *Report)           { rl.root = r }
func (rl *ReportLink) setParent(r ReportLinker)    { rl.parent = r }

func (r *Report) SetNode(link ReportLinker, uCredits int, kind any, name, path string) (ReportLinker, error) {
	var rl rootParenter

	ks, ok := kind.(string)
	if !ok {
		return nil, fmt.Errorf("Report SetNode: kind must be stringable, got: %T", kind)
	}

	if kk, ok := kind.(Kind); ok && kk.isMeter() { // this could also be implicit by excluding a name & path?
		rl = &ReportLeaf{Kind: kk.meterName(), UCreditUse: uCredits}
	} else {
		rl = &ReportNode{Kind: ks, UCreditSum: uCredits, Slug: name, Path: path}
	}

	rl.setRoot(r)
	rl.setParent(link)

	return rl.(ReportLinker), nil
}

func NewReporter() ReportLinker {
	report := &Report{}
	report.root = report
	return report
}
