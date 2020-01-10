// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package storebenchmark

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/store"
)

var reactionList []*model.Reaction

func BenchmarkReactionStore(b *testing.B, ss store.Store) {
	cases := []struct {
		PostId string `json:"post_id"`
	}{}
	data := getEnv("BENCHMARK_REACTIONS_DATA", `[{"post_id":"111fhwfzbbrzbe37zmsaqs3kcr"}]`)
	err := json.Unmarshal([]byte(data), &cases)
	if err != nil {
		b.Fatal("json deserialization error:", err)
		return
	}

	s := ss.Reaction()

	for idx, testcase := range cases {
		postId := testcase.PostId
		num := strconv.Itoa(idx + 1)
		b.Run("GetForPost/"+num, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				reactionList, appErr = s.GetForPost(postId, false)
			}
		})
	}
}
