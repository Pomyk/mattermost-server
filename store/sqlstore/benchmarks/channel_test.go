package benchmarks

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/mattermost/mattermost-server/model"
)

var channelList *model.ChannelList
var appErr *model.AppError

func BenchmarkChannels(b *testing.B) {
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
	for _, drv := range []string{"mysql", "postgres"} {
		supplier := setup(drv)
		store := supplier.Channel()
		for idx, testcase := range cases {
			teamId := testcase.TeamId
			userId := testcase.UserId
			term := testcase.Term
			num := strconv.Itoa(idx + 1)

			b.Run(drv+"/GetPublicChannelsForTeam/"+num, func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					channelList, appErr = store.GetPublicChannelsForTeam(teamId, 0, 100)
				}
			})
			b.Run(drv+"/GetMoreChannels/"+num, func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					channelList, appErr = store.GetMoreChannels(teamId, userId, 0, 100)
				}
			})
			b.Run(drv+"/GetDeleted/"+num, func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					channelList, appErr = store.GetDeleted(teamId, 0, 100, userId)
				}
			})
			b.Run(drv+"/SearchInTeam/"+num, func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					channelList, appErr = store.SearchInTeam(teamId, term, false)
				}
			})
			b.Run(drv+"/SearchArchivedInTeam/"+num, func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					channelList, appErr = store.SearchArchivedInTeam(teamId, term, userId)
				}
			})
			b.Run(drv+"/SearchMore/"+num, func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					channelList, appErr = store.SearchMore(userId, teamId, term)
				}
			})
			b.Run(drv+"/SearchGroupChannels/"+num, func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					channelList, appErr = store.SearchGroupChannels(userId, term)
				}
			})
		}
	}
}
