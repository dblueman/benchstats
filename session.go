package main

type Session struct {
   name       string
   benchmarks map[string]*Benchmark
}

func (s *Session) stats() {
   for _, benchmark := range s.benchmarks {
      benchmark.stats()
   }
}

func (s *Session) print() {
   for _, benchmark := range s.benchmarks {
      benchmark.print()
   }
}
