package action

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"math"
	"math/rand"
	"strconv"
	"time"

	tl "gitlab.mvk.com/go/vkgo/pkg/vktl/gen/tldonutSubscriptions"
	"gitlab.mvk.com/go/vkgo/projects/donut/shooter/internal/constants"
	"gitlab.mvk.com/go/vkgo/projects/donut/shooter/internal/stats"
	"gitlab.mvk.com/go/vkgo/projects/donut/shooter/internal/utils"
)

//go:embed get_content_access_input.json
var input []byte

var _contents []map[string]string

func init() {
	var contentsDTO [][]map[string]string

	_contents = make([]map[string]string, 0)

	err := json.Unmarshal(input, &contentsDTO)
	if err != nil {
		panic(err)
	}

	for _, content := range contentsDTO {
		_contents = append(_contents, content...)
	}
}

type GetContentAccess struct {
	*BaseAction
	client *tl.Client
	stats  stats.GetContentAccess
}

func NewGetContentAccess(actorID int64, rpcPath string, errorProbability float64) *GetContentAccess {
	client, err := CreateTLClient(actorID, rpcPath, constants.DonutSubscriptionTimeout, &tl.Client{})
	if err != nil {
		panic(err)
	}

	return &GetContentAccess{
		BaseAction: NewBaseAction("get_content_access", errorProbability),
		client:     client,
		stats:      stats.GetContentAccess{},
	}
}

func (a *GetContentAccess) PrintCurrentAndFlush(name string, period time.Duration) {
	a.stats.PrintCurrentAndFlush("get content access", period)
}

func (a *GetContentAccess) PrintTotal(name string, shootingDuration time.Duration) {
	a.stats.PrintTotal("get content access", shootingDuration)
}

func (a *GetContentAccess) Do() {
	req, err := getRandomGetContentAccessRequest()
	if err != nil {
		panic(err)
	}

	var resp tl.GetContentsAccessResponse
	err = a.client.GetContentsAccess(context.Background(), req, nil, &resp)
	a.HandleError(err, req, resp)

	if err != nil || len(resp.ContentsAccess) == 0 {
		a.stats.RecordFailure()
	} else {
		a.stats.RecordSuccess()
	}
}

func getRandomGetContentAccessRequest() (tl.GetContentsAccess, error) {
	var r tl.GetContentsAccess

	r.Lang = 0
	r.UserId = utils.GenerateRequestedUserId()
	r.Connection = tl.Connection{
		Ip: GetRandomIP(),
	}
	r.Source = GetRandomSource()
	cs, err := getRandomContents()
	if err != nil {
		return tl.GetContentsAccess{}, err
	}

	r.Contents = cs
	r.Device = GetRandomDevice()

	return r, nil
}

func getRandomContents() ([]tl.Content, error) {
	const maxSize = 30

	flat := make([]map[string]string, 0)
	for _, row := range _contents {
		flat = append(flat, row)
	}

	i := rand.Intn(maxSize) + 1

	var list []tl.Content

	for range i {
		elementIDx := rand.Intn(len(flat))
		oid, err := strconv.ParseInt(flat[elementIDx]["owner_id"], 10, 64)
		if err != nil {
			return nil, err
		}

		contentID := flat[elementIDx]["content_id"]
		tInt, err := strconv.ParseInt(flat[elementIDx]["content_type"], 10, 64)

		if tInt > math.MaxUint8 || tInt < 0 {
			return nil, errors.New("invalid content type")
		}

		ct, err := MarshalContentType(tInt)

		list = append(list, tl.Content{
			Id:      contentID,
			OwnerId: oid,
			Type:    ct,
		})
	}

	return list, nil
}

func MarshalContentType(t int64) (tl.ContentType, error) {
	switch t {
	case 1:
		return tl.ContentTypeVideo(), nil
	case 2:
		return tl.ContentTypeArticle(), nil
	case 3:
		return tl.ContentTypeChat(), nil
	case 4:
		return tl.ContentTypePost(), nil
	case 5:
		return tl.ContentTypePhoto(), nil
	case 6:
		return tl.ContentTypeVoting(), nil
	default:
		return tl.ContentType{}, errors.New("unknown content type")
	}
}

func GetRandomSource() tl.SubscriptionSource {
	x := rand.Intn(4)

	switch x {
	case 0:
		return tl.SubscriptionSourceDescription()
	case 1:
		return tl.SubscriptionSourceVideoPlaceholder()
	case 2:
		return tl.SubscriptionSourceFriendsTabBanner()
	case 3:
		return tl.SubscriptionSourcePostPaywall()
	default:
		panic("invalid source")
	}

}

func GetRandomDevice() tl.Device {
	var d int32

	switch rand.Intn(6) {
	case 0:
		d = 2274003
	case 1:
		d = 3087106
	case 2:
		d = 3140623
	case 3:
		d = 6287487
	case 4:
		d = 7879029
	case 5:
		d = 8202606
	default:
		panic("invalid device")
	}

	return tl.Device{
		AppId: d,
	}
}
