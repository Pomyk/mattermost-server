// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package benchmarks

import (
	"strings"
	"testing"

	"github.com/mattermost/mattermost-server/model"
)

var reactionList []*model.Reaction

func BenchmarkReactions(b *testing.B) {
	postId := getEnv("BENCHMARK_REACTIONS_DATA", "111fhwfzbbrzbe37zmsaqs3kcr")

	drivers := strings.Split(getEnv("BENCHMARK_DRIVERS", "postgres,mysql"), ",")
	for _, drv := range drivers {
		supplier := setup(drv)
		store := supplier.Reaction()

		b.Run(drv+"/GetForPost", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				reactionList, appErr = store.GetForPost(postId, false)
			}
		})
	}
}
