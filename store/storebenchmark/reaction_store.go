// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package storebenchmark

import (
	"testing"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/store"
)

var reactionList []*model.Reaction

func BenchmarkReactionStore(b *testing.B, ss store.Store) {
	postId := getEnv("BENCHMARK_REACTIONS_DATA", "111fhwfzbbrzbe37zmsaqs3kcr")

	s := ss.Reaction()

	b.Run("GetForPost", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			reactionList, appErr = s.GetForPost(postId, false)
		}
	})
}
