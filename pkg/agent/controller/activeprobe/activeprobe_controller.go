/*
Copyright 2021 The Everoute Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package activeprobe

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/contiv/ofnet/ofctrl"
	"github.com/everoute/everoute/pkg/agent/datapath"
	activeprobev1alph1 "github.com/everoute/everoute/pkg/apis/activeprobe/v1alpha1"
	"github.com/everoute/everoute/pkg/constants"
)

type ActiveprobeController struct {
	// k8sClient used to create/read/update activeprobe
	k8sClient client.Client
	Scheme    *runtime.Scheme

	DatapathManager *datapath.DpManager

	syncQueue workqueue.RateLimitingInterface
}

func NewActiveProbeController(client client.Client) *ActiveprobeController {
	return &ActiveprobeController{
		k8sClient: client,
		syncQueue: workqueue.NewRateLimitingQueue(workqueue.DefaultItemBasedRateLimiter()),
	}
}

func (a *ActiveprobeController) SetupWithManager(mgr ctrl.Manager) error {
	if mgr == nil {
		return fmt.Errorf("can't setup with nil manager")
	}

	var err error
	var activeProbeController controller.Controller

	if activeProbeController, err = controller.New("activeprobe_controller", mgr, controller.Options{
		MaxConcurrentReconciles: constants.DefaultMaxConcurrentReconciles,
		Reconciler:              reconcile.Func(a.ReconcileActiveProbe),
	}); err != nil {
		return err
	}

	if err = activeProbeController.Watch(&source.Kind{Type: &activeprobev1alph1.ActiveProbe{}}, &handler.Funcs{
		CreateFunc: a.AddActiveProbe,
		DeleteFunc: a.RemoveActiveProbe,
		UpdateFunc: a.UpdateActiveProbe,
	}); err != nil {
		return err
	}

	return nil
}

func (a *ActiveprobeController) HandlePacketIn(packetIn *ofctrl.PacketIn) {
	// In contoller runtime frame work, it's not easy to register packetIn callback in activeprobe controller
	// but we need active probe controller process packetIn for telemetry result parsing.
	// FIXME we need another module to parsing telemetry result and sync it to apiserver: update activeprobe status
}

func (a *ActiveprobeController) ReconcileActiveProbe(req ctrl.Request) (ctrl.Result, error) {
	// 1). received active probe request from work queue;
	// 2). parsing it;
	// 3). generate activeprobe flow rules and activeprobe packet
	// 4). inject probe packet
	// 5). (optional) register packet in handler in openflow controller
	ctx := context.Background()
	klog.V(2).Infof("ActiveprobeController received activeprobe %s reconcile", req.NamespacedName)

	ap := activeprobev1alph1.ActiveProbe{}
	if err := a.k8sClient.Get(ctx, req.NamespacedName, &ap); err != nil {
		klog.Errorf("unable to fetch activeprobe %v: %v", req.Name, err.Error())
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return ctrl.Result{}, nil
}

func (a *ActiveprobeController) Run(stopChan <-chan struct{}) {
	defer a.syncQueue.ShutDown()

	go wait.Until(a.SyncActiveProbeWorker, 0, stopChan)
	<-stopChan
}

func (a *ActiveprobeController) SyncActiveProbeWorker() {
	item, shutdown := a.syncQueue.Get()
	if shutdown {
		return
	}
	defer a.syncQueue.Done(item)

	objKey, ok := item.(k8stypes.NamespacedName)
	if !ok {
		a.syncQueue.Forget(item)
		klog.Errorf("Activeprobe %v was not found in workqueue", objKey)
		return
	}

	// TODO should support timeout and max retry
	if err := a.syncActiveProbe(objKey); err == nil {
		klog.Errorf("sync activeprobe  %v", objKey)
		a.syncQueue.Forget(item)
	} else {
		klog.Errorf("Failed to sync activeprobe %v, error: %v", objKey, err)
	}
}

func (a *ActiveprobeController) syncActiveProbe(objKey k8stypes.NamespacedName) error {
	var err error
	ctx := context.Background()
	ap := activeprobev1alph1.ActiveProbe{}
	if err := a.k8sClient.Get(ctx, objKey, &ap); err != nil {
		klog.Errorf("unable to fetch activeprobe %s: %s", objKey, err.Error())
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return client.IgnoreNotFound(err)
	}

	switch ap.Status.State {
	case activeprobev1alph1.ActiveProbeRunning:
		err = a.runActiveProbe(&ap)
	// TODO other state process
	case activeprobev1alph1.ActiveProbeCompleted:
	case activeprobev1alph1.ActiveProbeFailed:
	default:
	}

	return err
}

func (a *ActiveprobeController) runActiveProbe(ap *activeprobev1alph1.ActiveProbe) error {

	return nil
}

func (a *ActiveprobeController) ParseActiveProbeSpec() error {
	// Generate ip src && dst from endpoint name(or uuid), should store interface cache that contains ip

	return nil
}

func (a *ActiveprobeController) GenerateProbePacket() (*datapath.Packet, error) {

	return &datapath.Packet{}, nil
}

func (a *ActiveprobeController) SendActiveProbePacket() error {
	// Send activeprobe packet from the bridge which contains with src endpoint in probe spec
	// 1. ovsbr Name; 2. agent id
	var ovsbrName string
	var tag uint8
	var packet datapath.Packet
	var inPort, outPort uint32
	a.DatapathManager.SendActiveProbePacket(ovsbrName, packet, tag, inPort, &outPort)

	return nil
}

func (a *ActiveprobeController) InstallActiveProbeRuleFlow() error {
	return nil
}

func (a *ActiveprobeController) AddActiveProbe(e event.CreateEvent, q workqueue.RateLimitingInterface) {
	a.syncQueue.Add(ctrl.Request{NamespacedName: k8stypes.NamespacedName{
		Name:      e.Meta.GetName(),
		Namespace: e.Meta.GetNamespace(),
	}})
}

func (a *ActiveprobeController) UpdateActiveProbe(e event.UpdateEvent, q workqueue.RateLimitingInterface) {
	// should sync all object
	a.syncQueue.Add(ctrl.Request{NamespacedName: k8stypes.NamespacedName{
		Name:      e.MetaNew.GetName(),
		Namespace: e.MetaNew.GetNamespace(),
	}})
}

func (a *ActiveprobeController) RemoveActiveProbe(e event.DeleteEvent, q workqueue.RateLimitingInterface) {
	a.syncQueue.Add(ctrl.Request{NamespacedName: k8stypes.NamespacedName{
		Name:      e.Meta.GetName(),
		Namespace: e.Meta.GetNamespace(),
	}})
}
