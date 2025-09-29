package main

const (
	defaultSeed               int64   = 1
	defaultNumTrunks                  = 1001
	defaultAvgTrunkLen        float64 = 1
	defaultNumTypes                   = 15
	defaultAvgBranchLen       float64 = 3
	defaultBranchChildrenMean float64 = 5 // 1.7
	defaultFakeBranchLenMean  float64 = 2
	defaultDeadEndProb        float64 = 0.85
	defaultFakeBranchFac      float64 = 0.2 // max = 1
	defaultStartOnTrunkProb   float64 = 0.05
	defaultEndOnTrunkProb     float64 = 0.05
)

// -------------------------------
// Config
// -------------------------------

type Config struct {
	Seed int64

	// Магистраль
	NumTrunks     int       // число магистралей
	AvgTrunkLen   float64   // средняя длина магистрали
	TrunkTypeProb []float64 // распределение типов путей в магистрали

	// Переходы (склейки)
	NumTypes int
	TypeProb []float64 // распределение типов маршрутов
	Allow    [][]bool  // матрица возможных переходов (склеек)

	// Ветки деревьев от магистрали
	AvgBranchLen       float64 // общая длина ветки до конечной точки
	BranchChildrenMean float64 // сколько детей у каждой ветки

	// Фейковые ветки для нагрузки
	FakeBranchFac     float64
	FakeBranchLenMean float64
	DeadEndProb       float64

	// Точки на магистрали
	StartOnTrunkProb float64
	EndOnTrunkProb   float64
}

func defaultConfig() Config {
	return Config{
		Seed:               defaultSeed,
		NumTrunks:          defaultNumTrunks,
		AvgTrunkLen:        defaultAvgTrunkLen,
		NumTypes:           defaultNumTypes,
		AvgBranchLen:       defaultAvgBranchLen,
		BranchChildrenMean: defaultBranchChildrenMean,
		FakeBranchLenMean:  defaultFakeBranchLenMean,
		DeadEndProb:        defaultDeadEndProb,
		FakeBranchFac:      defaultFakeBranchFac,
		StartOnTrunkProb:   defaultStartOnTrunkProb,
		EndOnTrunkProb:     defaultEndOnTrunkProb,
	}
}
