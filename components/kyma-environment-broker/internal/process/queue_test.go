package process_test

import (
	"testing"
	"time"
	"github.com/sirupsen/logrus"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process/provisioning"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/provisioner"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process/provisioning/input"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/runtime"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/director"
	"github.com/stretchr/testify/require"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process"
	schema "github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
)

func TestQueue_Run(t *testing.T) {

	l := logrus.StandardLogger()
	memoryStorage := storage.NewMemoryStorage()
	repo := memoryStorage.Operations()
	iRepo := memoryStorage.Instances()
	mgr := process.NewManager(repo, l)
	provisionerClient := provisioner.NewFakeClient()

	runtimeProvider := runtime.NewComponentsListProvider("1.10.0", "managed-runtime-components.yaml")
	fullRuntimeComponentList, err := runtimeProvider.AllComponents()
	require.NoError(t, err)

	optionalComponentsDisablers := runtime.ComponentsDisablers{
		"Loki":       runtime.NewLokiDisabler(),
		"Kiali":      runtime.NewGenericComponentDisabler("kiali", "kyma-system"),
		"Jaeger":     runtime.NewGenericComponentDisabler("jaeger", "kyma-system"),
		"Monitoring": runtime.NewGenericComponentDisabler("monitoring", "kyma-system"),
	}

	optComponentsSvc := runtime.NewOptionalComponentsService(optionalComponentsDisablers)


	inputFactory := input.NewInputBuilderFactory(optComponentsSvc, fullRuntimeComponentList, input.Config{

	}, "1.10.0")


	// create and run queue, steps provisioning
	inputInitialisation := provisioning.NewInputInitialisationStep(repo, inputFactory, "http://dummy.com")

	runtimeStep := provisioning.NewCreateRuntimeStep(repo, iRepo, provisionerClient, internal.ServiceManagerOverride{
		URL: "http://dummy.com",
	})
	dCli := director.NewFakeDirectorClient()
	runtimeStatusStep := provisioning.NewRuntimeStatusStep(repo, iRepo, provisionerClient, dCli)

	mgr.InitStep(inputInitialisation)

	mgr.AddStep(1, runtimeStep)
	mgr.AddStep(2, runtimeStatusStep)

	q := process.NewQueue(mgr)

	sCh := make(chan struct{})
	q.Run(sCh)

	po, _ := internal.NewProvisioningOperation("i-001", internal.ProvisioningParameters{
		ServiceID: "47c9dcbf-ff30-448e-ab36-d3bad66ba281",
		PlanID: "ca6e5357-707f-4565-bbbd-b3ab732597c6",
	})

	repo.InsertProvisioningOperation(po)

	q.Add(po.ID)


	time.Sleep(3 * time.Minute)

	op, _ := repo.GetProvisioningOperationByID(po.ID)

	provisionerClient.SetOperation(op.ProvisionerOperationID, schema.OperationStatus{
		State: schema.OperationStateSucceeded,
	})
	time.Sleep(65 * time.Second)
	inst, _ := iRepo.GetByID(op.InstanceID)
	dCli.SetConsoleURL(inst.RuntimeID, "http://done.com")

	time.Sleep(5 * time.Minute)
	t.Fail()
}
