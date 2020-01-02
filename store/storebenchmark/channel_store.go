package storebenchmark

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/store"
)

var channelList *model.ChannelList
var appErr *model.AppError

func BenchmarkChannelStore(b *testing.B, ss store.Store) {
	cases := []struct {
		UserId string `json:"user_id"`
		TeamId string `json:"team_id"`
		Term   string `json:"term"`
	}{}
	data := getEnv("BENCHMARK_CHANNELS_DATA", `[{"user_id":"9fhuxkgy7tgeijm7iu564abpbh","team_id":"tzmakzz6n3rq3c3mpn9bwdaogo","term":"john"}]`)
	err := json.Unmarshal([]byte(data), &cases)
	if err != nil {
		b.Fatal(err)
		return
	}
	s := ss.Channel()
	for idx, testcase := range cases {
		teamId := testcase.TeamId
		userId := testcase.UserId
		term := testcase.Term
		num := strconv.Itoa(idx + 1)

		b.Run("GetPublicChannelsForTeam/"+num, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				channelList, appErr = s.GetPublicChannelsForTeam(teamId, 0, 100)
			}
		})
		b.Run("GetMoreChannels/"+num, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				channelList, appErr = s.GetMoreChannels(teamId, userId, 0, 100)
			}
		})
		b.Run("GetDeleted/"+num, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				channelList, appErr = s.GetDeleted(teamId, 0, 100, userId)
			}
		})
		b.Run("SearchInTeam/"+num, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				channelList, appErr = s.SearchInTeam(teamId, term, false)
			}
		})
		b.Run("SearchArchivedInTeam/"+num, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				channelList, appErr = s.SearchArchivedInTeam(teamId, term, userId)
			}
		})
		b.Run("SearchMore/"+num, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				channelList, appErr = s.SearchMore(userId, teamId, term)
			}
		})
		b.Run("SearchGroupChannels/"+num, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				channelList, appErr = s.SearchGroupChannels(userId, term)
			}
		})
	}
}
