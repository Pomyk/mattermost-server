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

var postList *model.PostList

func BenchmarkPostStore(b *testing.B, ss store.Store) {
	cases := []struct {
		UserId    string `json:"user_id"`
		TeamId    string `json:"team_id"`
		ChannelId string `json:"channel_id"`
		PostId    string `json:"post_id"`
		Term      string `json:"term"`
	}{}
	data := getEnv("BENCHMARK_POSTS_DATA", `
		[{"user_id":"twqzra8tttghxpjedrqq7855nc","team_id":"1xf9p4d4b3n9ixsjkcta5ooopy","channel_id":"d6a66pydbf87draqpfbo6xmnjw","post_id":"g7tudgm7abgedmi86j9dhjrzgh","term":"john"},
		 {"user_id":"ah6kmr3r1pf9ueca4ck9tbukqo","team_id":"3hpnkqu75j8zfmjqr5xb1crgao","channel_id":"ng657d7ciprpzyqt8d6oh6p4ro","post_id":"p818d7d3jbbgtyjwikcjysnc4c","term":"john"}]
		`)
	err := json.Unmarshal([]byte(data), &cases)
	if err != nil {
		b.Fatal("json deserialization error:", err)
		return
	}
	s := ss.Post()
	for idx, testcase := range cases {
		teamId := testcase.TeamId
		userId := testcase.UserId
		channelId := testcase.ChannelId
		postId := testcase.PostId
		num := strconv.Itoa(idx + 1)

		b.Run("GetFlaggedPostsForTeam/"+num, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				postList, appErr = s.GetFlaggedPostsForTeam(userId, teamId, 0, 100)
			}
		})
		b.Run("GetFlaggedPostsForChannel/"+num, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				postList, appErr = s.GetFlaggedPostsForChannel(userId, channelId, 0, 100)
			}
		})
		b.Run("GetPosts/"+num, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				postList, appErr = s.GetPosts(channelId, 0, 60, false)
			}
		})
		b.Run("GetPosts(cached)/"+num, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				postList, appErr = s.GetPosts(channelId, 0, 60, true)
			}
		})
		b.Run("GetPostsSince/"+num, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				postList, appErr = s.GetPostsSince(channelId, 1574630071540, false)
			}
		})
		b.Run("GetPostsSince(cached)/"+num, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				postList, appErr = s.GetPostsSince(channelId, 1574630071540, true)
			}
		})
		b.Run("GetPostsAfter/"+num, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				postList, appErr = s.GetPostsAfter(channelId, postId, 100, 0)
			}
		})
	}
}
