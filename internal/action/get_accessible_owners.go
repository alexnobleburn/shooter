package action

import (
	"context"
	"time"

	tl "gitlab.mvk.com/go/vkgo/pkg/vktl/gen/tldonutSubscriptions"
	"gitlab.mvk.com/go/vkgo/projects/donut/shooter/internal/constants"
	"gitlab.mvk.com/go/vkgo/projects/donut/shooter/internal/stats"
	"gitlab.mvk.com/go/vkgo/projects/donut/shooter/internal/utils"
)

type GetAccessibleOwners struct {
	*BaseAction
	client *tl.Client
	stats  stats.Stats
}

func NewGetAccessibleOwners(actorID int64, rpcPath string, errorProbability float64) *GetAccessibleOwners {
	client, err := CreateTLClient(actorID, rpcPath, constants.DonutSubscriptionTimeout, &tl.Client{})
	if err != nil {
		panic(err)
	}

	return &GetAccessibleOwners{
		BaseAction: NewBaseAction("get_accessible_owners", errorProbability),
		client:     client,
		stats:      stats.Stats{},
	}
}

func (a *GetAccessibleOwners) PrintCurrentAndFlush(name string, period time.Duration) {
	a.stats.PrintCurrentAndFlush("get accessible owners", period)
}

func (a *GetAccessibleOwners) PrintTotal(name string, shootingDuration time.Duration) {
	a.stats.PrintTotal("get accessible owners", shootingDuration)
}

func (a *GetAccessibleOwners) Do() {
	req := getRandomGetAccessibleOwnersRequest()

	var resp tl.GetAccessibleOwnersResponse
	err := a.client.GetAccessibleOwners(context.Background(), req, nil, &resp)
	a.HandleError(err, req, resp)

	if err != nil || len(resp.AccessibleOwnerIds) == 0 {
		a.stats.RecordFailure()
	} else {
		a.stats.RecordSuccess()
	}
}

func getRandomGetAccessibleOwnersRequest() tl.GetAccessibleOwners {
	return tl.GetAccessibleOwners{
		UserId: utils.GenerateRequestedUserId(),
	}
}
