package process

import (
	"testing"
	"time"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"golang.org/x/tools/go/cfg"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process/provisioning"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/provisioner"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process/provisioning/input"
)

type e struct {

}

func (e) Execute(opId string) (time.Duration, error) {
	fmt.Printf("---- %s\n", opId)
	panic(fmt.Errorf("hi there!"))
	return 5 * time.Second, nil
}

func TestQueue_Run(t *testing.T) {
	l := logrus.StandardLogger()
	memoryStorage := storage.NewMemoryStorage()
	repo := memoryStorage.Operations()
	mgr := NewManager(repo, l)
	provisionerClient := provisioner.NewFakeClient()

	inputFactory := input.NewInputBuilderFactory(optComponentsSvc, fullRuntimeComponentList, cfg.Provisioning, cfg.KymaVersion)


	// create and run queue, steps provisioning
	inputInitialisation := provisioning.NewInputInitialisationStep(repo, inputFactory, "http://dummy.com")

	runtimeStep := provisioning.NewCreateRuntimeStep(repo, provisionerClient, internal.ServiceManagerOverride{
		URL: "http://dummy.com",
	})
	runtimeStatusStep := provisioning.NewRuntimeStatusStep(repo, memoryStorage.Instances(), provisionerClient, directorClient)

	stepManager.InitStep(inputInitialisation)

	stepManager.AddStep(1, runtimeStep)
	stepManager.AddStep(2, runtimeStatusStep)

	q := NewQueue(mgr)

	sCh := make(chan struct{})
	q.Run(sCh)

	q.Add("001")

	go func() {
		for {
			fmt.Println(time.Now())
			time.Sleep(1 * time.Second)
		}
	}()
	time.Sleep(10 * time.Minute)

}
