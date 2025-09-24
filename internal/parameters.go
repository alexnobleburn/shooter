package internal

import (
	"flag"
	"fmt"
	"math"
	"os"
	"time"
)

type Parameters struct {
	Workers          int
	ShootingDuration time.Duration
	AvgTiming        time.Duration
	StatsPeriod      time.Duration
	RpcPath          string
	ActorID          int64
	ErrorProbability float64
}

func LoadParameters() Parameters {
	errorProbability := flag.Float64("ep", 0.001, "Вероятность записи ошибки в консоль и файл")
	targetRps := flag.Uint("target-rps", 0, "Приблизительная оценка нагрузки, > 0")
	shootingDuration := flag.Duration("shooting-duration", 0, "Продолжительность обстрела, > 0")
	avgTiming := flag.Duration("avg-timing", time.Duration(1)*time.Millisecond, "Приблизительная оценка таймингов одного запроса")
	statsPeriod := flag.Duration("stats-period", time.Duration(30)*time.Second, "Частота вывода статистики")

	defaultPath := fmt.Sprintf("unix:///var/kphp/%s/socket/vkontakte/2399", os.Getenv("USER"))
	rpcPath := flag.String("rpc-path", defaultPath, "path to the socket")
	actorID := flag.Int64("actor-id", 17142, "id of the actor")
	flag.Parse()

	if *targetRps <= 0 || *shootingDuration < 1*time.Second {
		fmt.Printf("error: invalid parameters\n")
		os.Exit(1)
	}

	if *errorProbability == 0.0 {
		fmt.Printf("error: error probability can't be zero\n")
		os.Exit(1)
	}

	workerPerformance := float64(1000) / float64(avgTiming.Milliseconds())
	workers := int(math.Ceil(float64(*targetRps) / workerPerformance))
	if workers == 0 {
		workers = 1
	}

	parameters := Parameters{
		Workers:          workers,
		ShootingDuration: *shootingDuration,
		AvgTiming:        *avgTiming,
		StatsPeriod:      *statsPeriod,
		RpcPath:          *rpcPath,
		ActorID:          *actorID,
		ErrorProbability: *errorProbability,
	}
	fmt.Printf("shooting parameters: %+v\n", parameters)
	return parameters
}
