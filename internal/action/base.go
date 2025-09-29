package action

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"gitlab.mvk.com/go/vkgo/pkg/paas/rpcf"
	"gitlab.mvk.com/go/vkgo/pkg/rpc"
	"gitlab.mvk.com/go/vkgo/projects/donut/shooter/internal/utils"
)

type BaseAction struct {
	fileLoader       *FileLoader
	errorProbability float64
	actionName       string
}

func NewBaseAction(actionName string, errorProbability float64) *BaseAction {
	fileName := fmt.Sprintf("%s_launch_%d.jsonl", actionName, time.Now().UnixNano())

	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		panic(err)
	}

	fileLoader, err := NewFileLoader(file)
	if err != nil {
		panic(err)
	}

	return &BaseAction{
		fileLoader:       fileLoader,
		errorProbability: errorProbability,
		actionName:       actionName,
	}
}

func (b *BaseAction) Close() error {
	return b.fileLoader.Close()
}

func (b *BaseAction) HandleError(err error, request interface{}, response interface{}) {
	if err != nil && rand.Float64() < b.errorProbability {
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
			ID         string      `json:"id"`
			Request    interface{} `json:"request"`
			Response   interface{} `json:"response"`
			StatusCode int32       `json:"status_code,omitempty"`
			ErrMessage string      `json:"err_message,omitempty"`
		}{
			ID:         id.String(),
			Request:    request,
			Response:   response,
			ErrMessage: err.Error(),
			StatusCode: status,
		}

		if err = b.fileLoader.Load(loaderData); err != nil {
			log.Printf("file loader: %s", err.Error())
		}
	}
}

type FileLoader struct {
	file *os.File
	mu   sync.Mutex
}

func NewFileLoader(file *os.File) (*FileLoader, error) {
	return &FileLoader{file: file}, nil
}

func (f *FileLoader) Close() error {
	return f.file.Close()
}

func (f *FileLoader) Load(data interface{}) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	b, err := json.Marshal(data)
	if err != nil {
		return err
	}

	b = append(b, '\n')
	if _, err = f.file.Write(b); err != nil {
		return err
	}

	return nil
}

func GetRandomIP() string {
	return fmt.Sprintf("%d.%d.%d.%d", rand.Intn(256), rand.Intn(256), rand.Intn(256), rand.Intn(256))
}

type TLClient interface {
	Call(ctx context.Context, method string, request interface{}, options interface{}, response interface{}) error
	GetTimeout() time.Duration
}

type TLAction struct {
	*BaseAction
	timeout time.Duration
}

func NewTLAction(actionName string, actorID int64, prcPath string, errorProbability float64, timeout time.Duration) *TLAction {
	return &TLAction{
		BaseAction: NewBaseAction(actionName, errorProbability),
		timeout:    timeout,
	}
}

func (ta *TLAction) GetTimeout() time.Duration {
	return ta.timeout
}

func CreateTLClient[T interface{}](actorID int64, rpcPath string, timeout time.Duration, client *T) (*T, error) {
	client, err := rpcf.TlClientFromManager(
		utils.InitRpcManager(rpcPath),
		&rpcf.ConnectionConfig{
			ActorID: actorID,
			Timeout: timeout,
		},
		client,
	)

	if err != nil {
		return nil, err
	}

	log.Printf("Timeout: %s\n", timeout.String())
	return client, nil
}
