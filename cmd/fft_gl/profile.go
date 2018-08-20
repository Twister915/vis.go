package main

import (
	"os"
	"runtime/pprof"
	"sync"

	"github.com/rs/zerolog/log"
)

func memoryProfileHandler(wg *sync.WaitGroup, shutdown <-chan struct{}) {
	defer wg.Done()

	dest := os.Getenv("MEM_PROFILE")
	if dest == "" {
		return
	}

	log.Info().Str("dest", dest).Msg("init memory profile")

	<-shutdown
	f, err := os.OpenFile(dest, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	if err = pprof.WriteHeapProfile(f); err != nil {
		panic(err)
	}
}

func cpuProfileHandler(wg *sync.WaitGroup, shutdown <-chan struct{}) {
	defer wg.Done()

	dest := os.Getenv("CPU_PROFILE")
	if dest == "" {
		return
	}

	log.Info().Str("dest", dest).Msg("init cpu profile")
	f, err := os.OpenFile(dest, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	if err = pprof.StartCPUProfile(f); err != nil {
		panic(err)
	}

	<-shutdown

	pprof.StopCPUProfile()
}
