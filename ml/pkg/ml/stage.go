package ml

//go:generate stringer -type=Stage -linecomment  -trimprefix Stage
//go:generate jsonenums -type=Stage
type Stage uint

const (
	StageDataPreparation     Stage = iota + 1 //data preparation
	StageDataSplitting                        //data splitting
	StageModelBuilding                        //model building
	StagePredicting                           //predicting
	StageQuantileCalculation                  //quantile calculation
)
