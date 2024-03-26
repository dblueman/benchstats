package main

import (
   "flag"
   "fmt"
   "os"
   "regexp"
   "strconv"
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

func compare(b1 *Benchmark, b2 *Benchmark) {
   fmt.Printf("%s: %5.1f σ%5.2f  %5.1f σ%4.2f  %4.1f%%\n", b1.name, b1.mean, b1.stdDev, b2.mean, b2.stdDev, 100 * b2.mean / b1.mean)
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

   fmt.Println("mean runtimes (s) - lower is better")
   for name, b1 := range sessions[0].benchmarks {
      b2, ok := sessions[1].benchmarks[name]

      if ok {
         compare(b1, b2)
      }
   }

   return nil
}

func main() {
   flag.Usage = func() {
      fmt.Fprintf(os.Stderr, "Usage: benchstats <NASA NPB baseline> [NASA NPB new]\n")
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
