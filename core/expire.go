package core

import (
	"log"
	"time"
)

func expireSample() float32 {

	var limit int = 20
	var expiredCount int = 0

	for key, value := range store {
		if value.ExpiresAt != -1 {
			limit--

			if value.ExpiresAt <= time.Now().UnixMilli() {
				delete(store, key)
				expiredCount++
			}
		}

		// once we reach 20 break the loop
		if limit == 0 {
			break
		}
	}

	return float32(expiredCount) / float32(limit)

}

func DeleteExpiredKeys() {

	for {
		var val = expireSample()
		if val < 0.25 {
			break
		}
	}
	log.Println("deleted the expired but undeleted keys. total keys : ", len(store))

}
