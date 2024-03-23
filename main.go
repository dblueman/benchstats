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

func parse(fname string) (* Session, error) {
   text, err := os.ReadFile(fname)
   if err != nil {
      return nil, fmt.Errorf("parse: %w", err)
   }

   session := Session{benchmarks: map[string]*Benchmark{}}
   matches := regexpNPB.FindAllSubmatch(text, -1)

   for _, match := range matches {
      name := string(match[1]) + "-" + string(match[2])
      runtime, err := strconv.ParseFloat(string(match[3]), 64)
      if err != nil {
         return nil, fmt.Errorf("parse: %w")
      }

      _, ok := session.benchmarks[name]
      if !ok {
         session.benchmarks[name] = &Benchmark{name: name, runtimes: []float64{}}
      }

      session.benchmarks[name].runtimes = append(session.benchmarks[name].runtimes, runtime)
   }

   return &session, nil
}

func top(infiles []string) error {
   for _, infile := range infiles {
      session, err := parse(infile)
      if err != nil {
         return err
      }

      session.stats()
      session.print()
   }

   return nil
}

func main() {
   flag.Usage = func() {
   	fmt.Fprintf(os.Stderr, "Usage: benchstats <summary>\n")
   	flag.PrintDefaults()
   }
   flag.Parse()

   if flag.NArg() != 1 {
      flag.Usage()
      os.Exit(2)
   }

   err := top(flag.Args())
   if err != nil {
      fmt.Fprintf(os.Stderr, err.Error())
      os.Exit(1)
   }
}
