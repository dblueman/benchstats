package main

import (
   "flag"
   "fmt"
   "math"
   "os"
   "regexp"
   "strconv"
)

const (
   Zscore95 = 1.959964
)

var (
   regexpNPB = regexp.MustCompile(`(?m)^ ([A-Z]{2}) Benchmark Completed.\n Class += +([A-E])\n Size.*\n Iterations.*\n Time in seconds =  +(\d+\.\d+)\n Total threads += +(\d+)$`)
)

func parse(fname string) (Session, error) {
   text, err := os.ReadFile(fname)
   if err != nil {
      return Session{}, fmt.Errorf("parse: %w", err)
   }

   session := Session{benchmarks: map[string]*Benchmark{}}
   matches := regexpNPB.FindAllSubmatch(text, -1)

   for _, match := range matches {
      name := string(match[1]) + "-" + string(match[2])
      runtime, err := strconv.ParseFloat(string(match[3]), 64)
      if err != nil {
         return Session{}, fmt.Errorf("parse: %w")
      }

      _, ok := session.benchmarks[name]
      if !ok {
         session.benchmarks[name] = &Benchmark{name: name, runtimes: []float64{}}
      }

      session.benchmarks[name].runtimes = append(session.benchmarks[name].runtimes, runtime)
   }

   return session, nil
}

func top(infiles []string) error {
   sessions := []Session{}

   for _, infile := range infiles {
      session, err := parse(infile)
      if err != nil {
         return err
      }

      session.stats()
      sessions = append(sessions, session)
   }

   if len(sessions) == 1 {
      sessions[0].print()
      return nil
   }

   fmt.Println("            runtime (s)")
   fmt.Println("    samples     A     B  diff  error")

   totalCount := 0
   totalDiff := 0.

   for name, b1 := range sessions[0].benchmarks {
      b2, ok := sessions[1].benchmarks[name]

      if !ok {
         continue
      }

      meanDiff := b1.mean - b2.mean
      meanDiffPercent := 100 * meanDiff / b1.mean
      marginOfError := Zscore95 * math.Sqrt(math.Pow(b1.stdDev, 2) / float64(len(b1.runtimes)) + math.Pow(b2.stdDev, 2) / float64(len(b2.runtimes)))

      fmt.Printf("%s: %5d %5.1f %5.1f %4.1f%% ± %3.2f\n", b1.name, len(b1.runtimes) + len(b2.runtimes), b1.mean, b2.mean, meanDiffPercent, marginOfError)
      totalCount++
      totalDiff += meanDiff
   }

   fmt.Printf(" avg:                   %4.1f%%\n", totalDiff / float64(totalCount))

   return nil
}

func main() {
   flag.Usage = func() {
      fmt.Fprintf(os.Stderr, "Usage: benchstats <NASA NPB output A> [NASA NPB output B]\n")
      flag.PrintDefaults()
   }
   flag.Parse()

   if flag.NArg() < 1 || flag.NArg() > 2 {
      flag.Usage()
      os.Exit(2)
   }

   err := top(flag.Args())
   if err != nil {
      fmt.Fprintf(os.Stderr, err.Error())
      os.Exit(1)
   }
}
