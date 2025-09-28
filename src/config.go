package main

const (
	defaultSeed               int64   = 1
	defaultNumTrunks                  = 2000
	defaultAvgTrunkLen        float64 = 3
	defaultNumTypes                   = 5
	defaultAvgBranchLen       float64 = 3
	defaultBranchChildrenMean float64 = 1.7
	defaultFakeBranchLenMean  float64 = 2
	defaultDeadEndProb        float64 = 0.85
	defaultFakeBranchFac      float64 = 0.3 // max = 1
	defaultStartOnTrunkProb   float64 = 0
	defaultEndOnTrunkProb     float64 = 0
	defaultMakePairs          bool    = true
)

// -------------------------------
// Config
// -------------------------------

type Config struct {
	Seed int64

	// Trunks
	NumTrunks     int
	AvgTrunkLen   float64
	TrunkLenMin   int
	TrunkTypeProb []float64

	// Types and transitions
	NumTypes int
	TypeProb []float64
	Allow    [][]bool

	// Feeder trees
	AvgBranchLen       float64
	BranchChildrenMean float64

	// Fake spikes
	FakeBranchFac float64
	SpikeLenMean  float64
	DeadEndProb   float64

	// Endpoints on trunk
	StartOnTrunkProb float64
	EndOnTrunkProb   float64

	// Pairs for benchmark
	MakePairs bool
}

func defaultConfig() Config {
	return Config{
		Seed:               defaultSeed,
		NumTrunks:          defaultNumTrunks,
		AvgTrunkLen:        defaultAvgTrunkLen,
		NumTypes:           defaultNumTypes,
		AvgBranchLen:       defaultAvgBranchLen,
		BranchChildrenMean: defaultBranchChildrenMean,
		SpikeLenMean:       defaultFakeBranchLenMean,
		DeadEndProb:        defaultDeadEndProb,
		FakeBranchFac:      defaultFakeBranchFac,
		StartOnTrunkProb:   defaultStartOnTrunkProb,
		EndOnTrunkProb:     defaultEndOnTrunkProb,
		MakePairs:          defaultMakePairs,
	}
}
