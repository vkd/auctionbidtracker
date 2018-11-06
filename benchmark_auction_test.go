package auctionbidtracker

import (
	"sync"
	"testing"
)

func Benchmark_MemoryAuction(b *testing.B) {
	benchmarkAuction(b, func() Auctioner { return NewMemoryAuction() })
}

func Benchmark_FastMemoryAuction(b *testing.B) {
	benchmarkAuction(b, func() Auctioner { return NewFastMemoryAuction() })
}

func benchmarkAuction(b *testing.B, fn func() Auctioner) {
	for i := 0; i < b.N; i++ {
		a := fn()

		var wg sync.WaitGroup
		for users := 0; users < 10; users++ {

			wg.Add(1)
			go func(users int) {

				item, err := a.AddItem(&Item{ID: ItemID(users)})
				if err != nil {
					b.Errorf("Error on add item: %v", err)
				}

				_ = item

				for bids := 0; bids < 1000; bids++ {
					err = a.MakeBid(item.ID, UserID(users), &Bid{Value: i})
					if err != nil {
						b.Errorf("Error on make bid: %v", err)
					}
				}
				wg.Done()

			}(users) // go func

		} // for users
		wg.Wait()

	} // for b.N
}
