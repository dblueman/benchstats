package main

import (
   "flag"
   "fmt"
   "math"
   "os"
   "regexp"
   "sort"
   "strconv"
)

type Result struct {
   name    string
   A, B    float64
   diff    float64
   err     float64
}

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

   session := Session{
      name:       fname,
      benchmarks: map[string]*Benchmark{},
   }
   matches := regexpNPB.FindAllSubmatch(text, -1)

   for _, match := range matches {
      name := string(match[1]) + "-" + string(match[2])
      runtime, err := strconv.ParseFloat(string(match[3]), 64)
      if err != nil {
         return Session{}, fmt.Errorf("parse: invalid number '%v'", match[3])
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

   fmt.Println("  runs description")
   for i, session := range sessions {
      runs := 0

      for _, benchmark := range session.benchmarks {
         runs += len(benchmark.runtimes)
      }

      fmt.Printf("%c %4d %s\n", 'A'+i, runs, session.name)
   }

   fmt.Println("\n        runtime (s)")
   fmt.Println("           A      B    diff    error")

   totalDiff := 0.
   results := []Result{}

   for name, b1 := range sessions[0].benchmarks {
      b2, ok := sessions[1].benchmarks[name]

      if !ok {
         continue
      }

      meanDiff := b1.mean - b2.mean
      marginOfError := Zscore95 * math.Sqrt(math.Pow(b1.stdDev, 2) / float64(len(b1.runtimes)) + math.Pow(b2.stdDev, 2) / float64(len(b2.runtimes)))
      result := Result{
         name: b1.name,
         A:       b1.mean,
         B:       b2.mean,
         diff:    100 * meanDiff / b1.mean,
         err:     marginOfError,
      }

      totalDiff += meanDiff
      results = append(results, result)
   }

   sort.Slice(results, func(i, j int) bool {
      return results[i].diff < results[j].diff
   })

   for _, result := range results {
      fmt.Printf("%s: %6.1f %6.1f %6.1f%% ± %6.2f\n", result.name, result.A, result.B, result.diff, result.err)
   }

   fmt.Printf(" avg:               %6.1f%%\n", totalDiff / float64(len(results)))

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
      fmt.Fprint(os.Stderr, err.Error())
      os.Exit(1)
   }
}
