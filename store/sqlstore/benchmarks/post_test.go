// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package benchmarks

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/mattermost/mattermost-server/v5/model"
)

var postList *model.PostList

func BenchmarkPosts(b *testing.B) {
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
		b.Fatal(err)
		return
	}
	drivers := strings.Split(getEnv("BENCHMARK_DRIVERS", "postgres,mysql"), ",")
	for _, drv := range drivers {
		supplier := setup(drv)
		store := supplier.Post()
		for idx, testcase := range cases {
			teamId := testcase.TeamId
			userId := testcase.UserId
			channelId := testcase.ChannelId
			postId := testcase.PostId
			num := strconv.Itoa(idx + 1)

			b.Run(drv+"/GetFlaggedPostsForTeam/"+num, func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					postList, appErr = store.GetFlaggedPostsForTeam(userId, teamId, 0, 100)
				}
			})
			b.Run(drv+"/GetFlaggedPostsForChannel/"+num, func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					postList, appErr = store.GetFlaggedPostsForChannel(userId, channelId, 0, 100)
				}
			})
			b.Run(drv+"/GetPosts(skipThreads=true)/"+num, func(b *testing.B) {
				opts := model.GetPostsOptions{
					ChannelId:        channelId,
					Page:             0,
					PerPage:          100,
					SkipFetchThreads: true,
				}
				for i := 0; i < b.N; i++ {
					postList, appErr = store.GetPosts(opts, false)
				}
			})
			b.Run(drv+"/GetPosts(skipThreads=false)/"+num, func(b *testing.B) {
				opts := model.GetPostsOptions{
					ChannelId:        channelId,
					Page:             0,
					PerPage:          100,
					SkipFetchThreads: false,
				}
				for i := 0; i < b.N; i++ {
					postList, appErr = store.GetPosts(opts, false)
				}
			})
			b.Run(drv+"/GetPostsSince(skipThreads=true)/"+num, func(b *testing.B) {
				opts := model.GetPostsSinceOptions{
					ChannelId:        channelId,
					Time:             1574630071540,
					SkipFetchThreads: true,
				}
				for i := 0; i < b.N; i++ {
					postList, appErr = store.GetPostsSince(opts, false)
				}
			})
			b.Run(drv+"/GetPostsSince(skipThreads=false)/"+num, func(b *testing.B) {
				opts := model.GetPostsSinceOptions{
					ChannelId:        channelId,
					Time:             1574630071540,
					SkipFetchThreads: false,
				}
				for i := 0; i < b.N; i++ {
					postList, appErr = store.GetPostsSince(opts, false)
				}
			})
			b.Run(drv+"/GetPostsAfter(skipThreads=true)/"+num, func(b *testing.B) {
				opts := model.GetPostsOptions{
					ChannelId:        channelId,
					PostId:           postId,
					SkipFetchThreads: true,
					Page:             0,
					PerPage:          100,
				}
				for i := 0; i < b.N; i++ {
					postList, appErr = store.GetPostsAfter(opts)
				}
			})
			b.Run(drv+"/GetPostsAfter(skipThreads=false)/"+num, func(b *testing.B) {
				opts := model.GetPostsOptions{
					ChannelId:        channelId,
					PostId:           postId,
					SkipFetchThreads: false,
					Page:             0,
					PerPage:          100,
				}
				for i := 0; i < b.N; i++ {
					postList, appErr = store.GetPostsAfter(opts)
				}
			})
		}
	}
}
