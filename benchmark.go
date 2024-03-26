package main

import (
   "fmt"
   "math"
   "os"

   "github.com/montanaflynn/stats"
)

type Benchmark struct {
   name     string
   runtimes []float64
   mean     float64
   stdDev   float64
}

func (b *Benchmark) stats() {
   var err error
   b.mean, err = stats.Mean(b.runtimes)
   if err != nil {
      fmt.Fprintln(os.Stderr, err.Error())
   }

   b.stdDev, err = stats.StandardDeviation(b.runtimes)
   if err != nil {
      fmt.Fprintln(os.Stderr, err.Error())
   }
}

func (b *Benchmark) print() {
   fmt.Printf("%s:", b.name)

   if b.mean != 0. && b.stdDev != 0. {
      marginOfError := Zscore95 * (b.stdDev / math.Sqrt(float64(len(b.runtimes))))
      fmt.Printf(" %6.2f ± %2.3f (%d samples)", b.mean, marginOfError, len(b.runtimes))
   } else {
      for _, runtime := range b.runtimes {
         fmt.Printf(" %.2f", runtime)
      }
   }

   fmt.Println()
}
