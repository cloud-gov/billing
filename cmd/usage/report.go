package main

import (
	"regexp"
)

type ReportWriter interface {
	SetNode(link ReportLinker, uCredits int, kind any, name, path string) (ReportLinker, error)
}

type ReportLinker interface {
	getParent() ReportLinker
	getChildren() []ReportLinker
	addChild(ReportLinker)
}

type rootParenter interface {
	setRoot(r *Report)
	setParent(r ReportLinker)
}

func sliceToReportLinkers[S ~[]N, N ReportLinker](s S) (l []ReportLinker) {
	for _, n := range s {
		l = append(l, n)
	}
	return l
}

type Kind string

const (
	Org   Kind = "cf_org"
	Space Kind = "cf_space"
	CfApp Kind = "meter::cfapps"
	CfSvc Kind = "meter::cfservices"
)

var meterReg = regexp.MustCompile(`^meter::(\w+)`)

func (k Kind) String() string {
	return string(k)
}

func (k Kind) isMeter() bool {
	res := meterReg.MatchString(k.String())
	return res
}

func (k Kind) meterName() string {
	return meterReg.FindStringSubmatch(k.String())[0]
}
