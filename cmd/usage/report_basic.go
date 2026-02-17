package main

import "fmt"

type (
	BasicReport struct {
		Orgs []*BasicReportOrg
		BasicReportLinker
	}
	BasicReportOrg struct {
		Name   string
		Spaces []*BasicReportSpace
		BasicReportLine
		BasicReportLinker
	}
	BasicReportSpace struct {
		Space string
		org   *BasicReportOrg
		BasicReportLine
		BasicReportLinker
	}
	BasicReportLine struct {
		UCreditsUtilized int
	}
	BasicReportLinker struct {
		root *BasicReport
	}
)

func (r *BasicReport) getChildren() []ReportLinker    { return sliceToReportLinkers(r.Orgs) }
func (o *BasicReportOrg) getChildren() []ReportLinker { return sliceToReportLinkers(o.Spaces) }
func (s *BasicReportSpace) getParent() ReportLinker   { return s.org }

func (l *BasicReportLinker) getParent() ReportLinker     { return l.root }
func (l *BasicReportLinker) getChildren() []ReportLinker { return nil }
func (l *BasicReportLinker) addChild(ReportLinker)       {}

func (r *BasicReport) SetOrg(uCredits int, name string) (ReportLinker, error) {
	org := &BasicReportOrg{Name: name}
	org.root = r
	org.UCreditsUtilized = uCredits
	r.Orgs = append(r.Orgs, org)
	return org, nil
}

func (r *BasicReport) SetSpace(linker ReportLinker, uCredits int, name string) (ReportLinker, error) {
	org, ok := linker.(*BasicReportOrg)
	if !ok {
		return nil, fmt.Errorf("SetOrgSpaces: linker must be BasicReportOrg, got %T", linker)
	}

	space := &BasicReportSpace{Space: name, org: org}
	space.root = r
	space.UCreditsUtilized = uCredits

	org.Spaces = append(org.Spaces, space)

	return space, nil
}

func (r *BasicReport) SetNode(link ReportLinker, uCredits int, kind any, name, _ string) (ReportLinker, error) {
	if kind, ok := kind.(Kind); !ok {
		return nil, fmt.Errorf("BasicReport SetNode: kind must be Kind, got %T", kind)
	} else if kind == Org {
		return r.SetOrg(uCredits, name)
	} else {
		return r.SetSpace(link, uCredits, name)
	}
}

func NewBasicReporter() ReportLinker {
	report := &BasicReport{}
	report.root = report
	return report
}
