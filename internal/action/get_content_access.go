package action

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"gitlab.mvk.com/go/vkgo/pkg/paas/meowdb"
	"gitlab.mvk.com/go/vkgo/pkg/paas/rpcf"
	"gitlab.mvk.com/go/vkgo/pkg/rpc"
	tl "gitlab.mvk.com/go/vkgo/pkg/vktl/gen/tlDonutSubscriptions"
	"gitlab.mvk.com/go/vkgo/projects/donut/shooter/internal/stats"
	"gitlab.mvk.com/go/vkgo/projects/donut/shooter/internal/utils"
	"log"
	"math"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	donutSubscriptionTimeout = 1000 * time.Millisecond
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

type FileLoader struct {
	file *os.File
	mu   sync.Mutex
}

func NewFileLoader(file *os.File) (*FileLoader, error) {
	fl := &FileLoader{
		file: file,
	}

	return fl, nil
}

func (l *FileLoader) Close() error {
	return l.file.Close()
}

func (l *FileLoader) Load(data any) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	b = append(b, '\n')

	if _, err = l.file.Write(b); err != nil {
		return err
	}

	return nil
}

type DB struct {
	clusterSize int64
	dbClient    meowdb.Client
}

type GetContentAccess struct {
	fileLoader       *FileLoader
	db               *DB
	client           *tl.Client
	stats            stats.GetContentAccess
	errorProbability float64
}

func NewGetContentAccess(actorID int64, rpcPath string, errorProbability float64) *GetContentAccess {
	client, err := rpcf.TlClientFromManager(
		utils.InitRpcManager(rpcPath),
		&rpcf.ConnectionConfig{
			ActorID: actorID,
			Timeout: donutSubscriptionTimeout,
		},
		&tl.Client{},
	)

	if err != nil {
		panic(err)
	}

	log.Printf("Timeout: %s\n", client.Timeout.String())

	fileName := fmt.Sprintf("%s_%d.jsonl", "get_content_access_launch", time.Now().UnixNano())

	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		panic(err)
	}

	fileLoader, err := NewFileLoader(file)
	if err != nil {
		panic(err)
	}

	return &GetContentAccess{
		client:           client,
		stats:            stats.GetContentAccess{},
		fileLoader:       fileLoader,
		errorProbability: errorProbability,
	}
}

func (a *GetContentAccess) Close() error {
	return a.fileLoader.Close()
}

func (a *GetContentAccess) PrintCurrentAndFlush(name string, period time.Duration) {
	a.stats.PrintCurrentAndFlush(name, period)
}

func (a *GetContentAccess) PrintTotal(name string, shootingDuration time.Duration) {
	a.stats.PrintTotal(name, shootingDuration)
}

func (a *GetContentAccess) Do() {
	req, err := getRandomGetContentAccessRequest()
	if err != nil {
		panic(err)
	}

	var resp tl.GetContentsAccessResponse
	err = a.client.GetContentsAccess(context.Background(), req, nil, &resp)
	stat := a.handleStats(resp, err)

	if err != nil {
		if rand.Float64() < a.errorProbability {
			id := uuid.New()

			fmt.Printf("[%s] Ошибка при выполнении запроса: %v\n", id.String(), err)

			var status int32

			var val *rpc.Error
			if errors.As(err, &val) {
				status = val.Code

				//if strings.Contains(val.Error(), "invalid number of levels") {
				//	a.handleInvalidNumberOfLevels(req)
				//}
			}

			loaderData := struct {
				ID         string                       `json:"id"`
				Request    tl.GetContentsAccess         `json:"request"`
				Response   tl.GetContentsAccessResponse `json:"response"`
				StatusCode int32                        `json:"status_code,omitempty"`
				ErrMessage string                       `json:"err_message,omitempty"`
			}{
				ID:         id.String(),
				Request:    req,
				Response:   resp,
				ErrMessage: err.Error(),
				StatusCode: status,
			}

			if err = a.fileLoader.Load(loaderData); err != nil {
				log.Printf("file loader: %s", err.Error())
			}

		}
	}

	a.stats.Merge(stat)
}

//func (a *GetContentAccess) handleInvalidNumberOfLevels(req tl.GetContentsAccess) {
//	c := req.Contents
//
//	hash := make(map[types.OwnerID]map[types.ContentType][]types.ContentID)
//
//	for _, v := range c {
//		content, err := converter.NewContent().FromTL(v)
//		if err != nil {
//			panic(err)
//		}
//
//		if hash[content.ContentFullID.OwnerID] == nil {
//			hash[content.ContentFullID.OwnerID] = make(map[types.ContentType][]types.ContentID)
//		}
//
//		hash[content.ContentFullID.OwnerID][content.Type] = append(hash[content.ContentFullID.OwnerID][content.Type], content.ContentFullID.ContentID)
//	}
//}

func (a *GetContentAccess) handleStats(resp tl.GetContentsAccessResponse, err error) stats.GetContentAccess {
	var newStats stats.GetContentAccess
	if err != nil {
		newStats.GetContentAccessFail = 1
		newStats.Stats.Fail = 1
		return newStats
	}

	if len(resp.ContentsAccess) == 0 {
		newStats.GetContentAccessFail = 1
		newStats.Stats.Fail = 1
	} else {
		newStats.GetContentAccessSuccess = 1
		newStats.Stats.Success = 1
	}

	return newStats
}

func GetRandomIP() string {
	return fmt.Sprintf("%d.%d.%d.%d", rand.Intn(256), rand.Intn(256), rand.Intn(256), rand.Intn(256))
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
