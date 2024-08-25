package main

import (
	"context"
	"fmt"
	"github.com/sourcegraph/conc/pool"
	"time"
)

func main() {
	ctx, cancelFn := context.WithTimeout(context.Background(), 400*time.Millisecond)
	defer cancelFn()
	start := time.Now()
	p := pool.NewWithResults[int]().WithMaxGoroutines(10).WithErrors().WithContext(ctx)
	for i := 1; i <= 100; i++ {
		p.Go(func(ctx context.Context) (int, error) {
			select {
			case <-ctx.Done():
				return 0, ctx.Err()
			default:
				time.Sleep(100 * time.Millisecond)
				return i, nil
			}
		})
	}
	ints, err := p.Wait()
	for _, i := range ints {
		fmt.Printf("%d,", i)
	}
	fmt.Println("")

	errs := unwrapErrs(unwrapErr(err))
	fmt.Println("Total errors:", len(errs))

	fmt.Printf("Time elapsed: %v, size = %v\n", time.Since(start), len(ints))
}

func unwrapErrs(errs []error) []error {
	if len(errs) == 0 {
		return nil
	}
	if len(errs) == 1 {
		return unwrapErr(errs[0])
	}
	results := make([]error, 0)
	for _, e := range errs {
		results = append(results, unwrapErrs(unwrapErr(e))...)
	}
	return results
}

func unwrapErr(err error) []error {
	if err == nil {
		return nil
	}
	if uw, ok := err.(interface{ Unwrap() []error }); ok {
		return uw.Unwrap()
	}
	return []error{err}
}
