package utils

import (
	"math/rand"
	"slices"
)

const (
	groupProbability2 = 0.34
	groupProbability3 = 0.1
	groupProbability4 = 0.05
	groupProbability5 = 0.01

	over100kGroupProbability = 0.33

	preservedUserProbability = 0.1
	anonymousUserProbability = 0.001
)

var (
	groupIdLowerBound = 100_000_000
	groupIdUpperBound = 250_000_000

	over100kGroupIds = []int64{
		57846937, 23148107, 60302983, 50883936, 214133814,
		202136297, 207050983, 104523591, 147782201, 122560283,
		198071571, 166216692, 220829011, 155153648, 178132902,
		147286578, 118643747, 211504825, 182580928, 226944890,
	}

	userIdLowerBound = 100_000_000
	userIdUpperBound = 1_000_000_000

	preservedUserIds = []int64{
		1041624815, 111380812, 1027137966, 1033601908, 103512693,
	}
)

func GenerateRequestedGroupIds() []int64 {
	sliceLen := decideRequestedGroupIdsSliceLen()
	groupIds := make([]int64, 0, sliceLen)
	for i := 0; i < sliceLen; i++ {
		n := rand.Float64()
		var groupId int64
		if n < over100kGroupProbability {
			groupId = getRandomGroupOver100kId()
		} else {
			groupId = getRandomGroupId()
		}

		for slices.Contains(groupIds, groupId) {
			groupId = getRandomGroupId()
		}
		groupIds = append(groupIds, groupId)
	}
	return groupIds
}

func GenerateRequestedUserId() int64 {
	n := rand.Float64()
	if n < anonymousUserProbability {
		return 0
	}
	if n < preservedUserProbability {
		return getRandomPreservedUserId()
	}
	return getRandomUserId()
}

func decideRequestedGroupIdsSliceLen() int {
	n := rand.Float64()
	if n < groupProbability5 {
		return 5
	}
	if n < groupProbability4 {
		return 4
	}
	if n < groupProbability3 {
		return 3
	}
	if n < groupProbability2 {
		return 2
	}
	return 1
}

func getRandomGroupId() int64 {
	return int64(rand.Intn(groupIdUpperBound-groupIdLowerBound+1) + groupIdLowerBound)
}

func getRandomGroupOver100kId() int64 {
	return over100kGroupIds[rand.Intn(len(over100kGroupIds))]
}

func getRandomUserId() int64 {
	return int64(rand.Intn(userIdUpperBound-userIdLowerBound+1) + userIdLowerBound)
}

func getRandomPreservedUserId() int64 {
	return preservedUserIds[rand.Intn(len(preservedUserIds))]
}
