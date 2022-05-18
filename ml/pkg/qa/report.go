package qa

import "github.com/repofuel/repofuel/pkg/metrics"

type ModelReport struct {
	Format            string                `json:"format"             bson:"format"`
	Params            HyperParameters       `json:"params"             bson:"params,omitempty"`
	FeatureImportance map[string]float32    `json:"feature_importance" bson:"feature_importance,omitempty"`
	Accuracy          float32               `json:"accuracy"           bson:"accuracy"`
	MedianScore       float32               `json:"medians_score"      bson:"medians_score"`
	Buggy             ClassificationMetrics `json:"buggy"              bson:"buggy,omitempty"`
	Clean             ClassificationMetrics `json:"clean"              bson:"clean,omitempty"`
	WeightedAvg       ClassificationMetrics `json:"weighted_avg"       bson:"weighted_avg,omitempty"`
	MacroAvg          ClassificationMetrics `json:"macro_avg"          bson:"macro_avg,omitempty"`
}

type ModelMedians struct {
	All   metrics.ChangeMeasures `json:"all"    bson:"all"`
	Buggy metrics.ChangeMeasures `json:"buggy"  bson:"buggy"`
	Clean metrics.ChangeMeasures `json:"clean"  bson:"clean"`
}

type DataPoints struct {
	All     int      `json:"all"      bson:"all"`
	Train   TagsStat `json:"train"    bson:"train"`
	Test    TagsStat `json:"test"     bson:"test"`
	Predict int      `json:"predict"  bson:"predict"`
}

type TagsStat struct {
	NumBuggy int `json:"True"      bson:"buggy"`
	NumClean int `json:"False"     bson:"clean"`
}

type HyperParameters struct {
	Estimators  int    `json:"n_estimators"  bson:"n_estimators"`
	MaxFeatures string `json:"max_features"  bson:"max_features"`
	MaxDepth    int    `json:"max_depth"     bson:"max_depth"`
	Bootstrap   bool   `json:"bootstrap"     bson:"bootstrap"`
}

type ClassificationMetrics struct {
	F1Score   float32 `json:"f1-score"   bson:"f1_score"`
	Precision float32 `json:"precision"  bson:"precision"`
	Recall    float32 `json:"recall"     bson:"recall"`
	Support   int     `json:"support"    bson:"support"`
}
