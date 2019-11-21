package benchmarks

import (
	"testing"

	"github.com/mattermost/mattermost-server/model"
)

var reactionList []*model.Reaction

func BenchmarkReactions(b *testing.B) {
	postId := getEnv("BENCHMARK_REACTIONS_DATA", "111fhwfzbbrzbe37zmsaqs3kcr")

	for _, drv := range []string{"mysql", "postgres"} {
		supplier := setup(drv)
		store := supplier.Reaction()

		b.Run(drv+"/GetForPost", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				reactionList, appErr = store.GetForPost(postId, false)
			}
		})
	}
}
