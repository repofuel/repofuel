package insightgen

import (
	"fmt"

	"github.com/repofuel/repofuel/ingest/internal/entity"
	. "github.com/repofuel/repofuel/ingest/pkg/insights"
	"github.com/repofuel/repofuel/pkg/metrics"
)

const _Percentile50 = "0.5"
const _Percentile75 = "0.75"
const _Percentile90 = "0.9"

type quantileError struct {
	field      string
	percentile string
}

func (err quantileError) Error() string {
	if err.field == "" {
		return "missing quantile data"
	}

	return fmt.Sprintf("cannot find the precentil %s for %s", err.percentile, err.field)
}

type Generator struct {
	fileTop25   *metrics.FileMeasures
	fileTop10   *metrics.FileMeasures
	devTop25    *metrics.ChangeMeasures
	commitTop10 *metrics.ChangeMeasures
}

func NewGenerator(quantiles *metrics.Quantiles) (*Generator, error) {
	if quantiles == nil {
		return nil, quantileError{}
	}

	devTop25, ok := quantiles.Developer[_Percentile75]
	if !ok {
		return nil, quantileError{field: "developer", percentile: _Percentile75}
	}

	fileTop25, ok := quantiles.File[_Percentile75]
	if !ok {
		return nil, quantileError{field: "file", percentile: _Percentile75}
	}

	fileTop10, ok := quantiles.File[_Percentile90]
	if !ok {
		return nil, quantileError{field: "file", percentile: _Percentile75}
	}

	commitTop10, ok := quantiles.Commit[_Percentile90]
	if !ok {
		return nil, quantileError{field: "commit", percentile: _Percentile90}
	}

	return &Generator{
		fileTop25:   fileTop25,
		fileTop10:   fileTop10,
		devTop25:    devTop25,
		commitTop10: commitTop10,
	}, nil
}

func (gen *Generator) CommitInsights(c *entity.Commit) []Reason {
	if c == nil || c.Metrics == nil {
		return nil
	}

	var res []Reason

	res = gen.appendExpInsights(res, c.Metrics)
	res = gen.appendSizeInsights(res, c.Metrics)
	res = gen.appendDiffusionInsights(res, c.Metrics)

	return res
}

func (gen *Generator) FileInsights(c *entity.Commit) [][]Reason {
	if c == nil || c.Metrics == nil || len(c.Files) == 0 {
		return nil
	}

	res := make([][]Reason, len(c.Files))

	res = gen.appendFileHistoryInsights(res, c.Metrics, c.Files)
	res = gen.appendFileExpInsights(res, c.Metrics, c.Files)

	return res
}

func (gen *Generator) appendSizeInsights(res []Reason, m *metrics.ChangeMeasures) []Reason {

	if m.LT > gen.commitTop10.LT && m.LT > 300 {
		res = append(res, CommitLargeFiles)
	}

	if m.LA > gen.commitTop10.LA && m.LD > gen.commitTop10.LD && m.LA+m.LD > 100 {
		res = append(res, CommitIsLargest)

	} else if m.LA > gen.commitTop10.LA && m.LA > 50 {
		res = append(res, CommitIsLargestAdd)

	} else if m.LD > gen.commitTop10.LD && m.LD > 50 {
		res = append(res, CommitIsLargestDelete)
	}

	return res
}

func (gen *Generator) appendFileHistoryInsights(res [][]Reason, m *metrics.ChangeMeasures, files []*entity.File) [][]Reason {
	if m.NUC == 0 {
		return res
	}

	for i, f := range files {

		if f.Metrics.NFC > gen.fileTop25.NFC && f.Metrics.NFC > 10 {
			res[i] = append(res[i], FilesFrequentlyFixed)

		} else if f.Metrics.NUC > gen.fileTop25.NUC && f.Metrics.NUC > 20 {
			res[i] = append(res[i], FilesFrequentlyChanged)
		}

		if f.Metrics.AGE < 2 /* days */ && f.Metrics.AGE > 0 {
			res[i] = append(res[i], FilesRecentlyModified)

		} else if f.Metrics.AGE > gen.fileTop10.AGE && f.Metrics.AGE > 15 /* days */ {
			res[i] = append(res[i], FilesAbandoned)
		}
	}

	return res
}

func (gen *Generator) appendDiffusionInsights(res []Reason, m *metrics.ChangeMeasures) []Reason {

	if m.NS > gen.commitTop10.NS || m.ND > gen.commitTop10.ND || m.NF > gen.commitTop10.NF {
		res = append(res, CommitScattered)
	}

	return res
}

func (gen *Generator) appendExpInsights(res []Reason, m *metrics.ChangeMeasures) []Reason {
	if gen.fileTop25.EXP < 10 {
		return res
	}

	if m.EXP == 0 {
		return append(res, DevFirstCommit)
	}

	if gen.devTop25.EXP < 30 {
		return res
	}

	if m.EXP < 5 {
		return append(res, DevNewToRepository)
	}

	if m.EXP > gen.devTop25.EXP {
		res = append(res, DevExpert)
	}

	if m.SEXP < 10 && gen.commitTop10.SEXP > 10 {
		res = append(res, DevNewToSubsystem)
	}

	return res
}

func (gen *Generator) appendFileExpInsights(res [][]Reason, m *metrics.ChangeMeasures, files []*entity.File) [][]Reason {
	if gen.devTop25.EXP < 30 {
		return res
	}

	for i, f := range files {
		if f.Metrics.EXP > gen.fileTop25.EXP {
			res[i] = append(res[i], FilesDevExpert)

		} else if f.Metrics.EXP < 5 && f.Metrics.NUC > 5 {
			res[i] = append(res[i], FilesDevNew)
		}

		if f.SameDeveloper {
			res[i] = append(res[i], DevIsLastModifier)
		}
	}

	return res
}
