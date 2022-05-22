/*
Copyright 2014 Google Inc. All rights reserved.

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
package registry

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/coreos/go-etcd/etcd"
	. "github.com/pedrogao/simplek8s/pkg/api"
	. "github.com/pedrogao/simplek8s/pkg/client"
	"github.com/pedrogao/simplek8s/pkg/util"
)

// TODO: Move this to a common place, it's needed in multiple tests.
var apiPath = "/api/v1beta1"

func makeUrl(suffix string) string {
	return apiPath + suffix
}

type FakeTaskControl struct {
	controllerSpec []ReplicationController
	deleteTaskID   []string
}

func (f *FakeTaskControl) createReplica(spec ReplicationController) {
	f.controllerSpec = append(f.controllerSpec, spec)
}

func (f *FakeTaskControl) deleteTask(taskID string) error {
	f.deleteTaskID = append(f.deleteTaskID, taskID)
	return nil
}

func makeReplicationController(replicas int) ReplicationController {
	return ReplicationController{
		DesiredState: ReplicationControllerState{
			Replicas: replicas,
			TaskTemplate: TaskTemplate{
				DesiredState: TaskState{
					Manifest: ContainerManifest{
						Containers: []Container{
							Container{
								Image: "foo/bar",
							},
						},
					},
				},
				Labels: map[string]string{
					"name": "foo",
					"type": "production",
				},
			},
		},
	}
}

func makeTaskList(count int) TaskList {
	tasks := []Task{}
	for i := 0; i < count; i++ {
		tasks = append(tasks, Task{
			JSONBase: JSONBase{
				ID: fmt.Sprintf("task%d", i),
			},
		})
	}
	return TaskList{
		Items: tasks,
	}
}

func validateSyncReplication(t *testing.T, fakeTaskControl *FakeTaskControl, expectedCreates, expectedDeletes int) {
	if len(fakeTaskControl.controllerSpec) != expectedCreates {
		t.Errorf("Unexpected number of creates.  Expected %d, saw %d\n", expectedCreates, len(fakeTaskControl.controllerSpec))
	}
	if len(fakeTaskControl.deleteTaskID) != expectedDeletes {
		t.Errorf("Unexpected number of deletes.  Expected %d, saw %d\n", expectedDeletes, len(fakeTaskControl.deleteTaskID))
	}
}

func TestSyncReplicationControllerDoesNothing(t *testing.T) {
	body, _ := json.Marshal(makeTaskList(2))
	fakeHandler := util.FakeHandler{
		StatusCode:   200,
		ResponseBody: string(body),
	}
	testServer := httptest.NewTLSServer(&fakeHandler)
	client := Client{
		Host: testServer.URL,
	}

	fakeTaskControl := FakeTaskControl{}

	manager := MakeReplicationManager(nil, &client)
	manager.taskControl = &fakeTaskControl

	controllerSpec := makeReplicationController(2)

	manager.syncReplicationController(controllerSpec)
	validateSyncReplication(t, &fakeTaskControl, 0, 0)
}

func TestSyncReplicationControllerDeletes(t *testing.T) {
	body, _ := json.Marshal(makeTaskList(2))
	fakeHandler := util.FakeHandler{
		StatusCode:   200,
		ResponseBody: string(body),
	}
	testServer := httptest.NewTLSServer(&fakeHandler)
	client := Client{
		Host: testServer.URL,
	}

	fakeTaskControl := FakeTaskControl{}

	manager := MakeReplicationManager(nil, &client)
	manager.taskControl = &fakeTaskControl

	controllerSpec := makeReplicationController(1)

	manager.syncReplicationController(controllerSpec)
	validateSyncReplication(t, &fakeTaskControl, 0, 1)
}

func TestSyncReplicationControllerCreates(t *testing.T) {
	body := "{ \"items\": [] }"
	fakeHandler := util.FakeHandler{
		StatusCode:   200,
		ResponseBody: string(body),
	}
	testServer := httptest.NewTLSServer(&fakeHandler)
	client := Client{
		Host: testServer.URL,
	}

	fakeTaskControl := FakeTaskControl{}

	manager := MakeReplicationManager(nil, &client)
	manager.taskControl = &fakeTaskControl

	controllerSpec := makeReplicationController(2)

	manager.syncReplicationController(controllerSpec)
	validateSyncReplication(t, &fakeTaskControl, 2, 0)
}

func TestCreateReplica(t *testing.T) {
	body := "{}"
	fakeHandler := util.FakeHandler{
		StatusCode:   200,
		ResponseBody: string(body),
	}
	testServer := httptest.NewTLSServer(&fakeHandler)
	client := Client{
		Host: testServer.URL,
	}

	taskControl := RealTaskControl{
		kubeClient: client,
	}

	controllerSpec := ReplicationController{
		DesiredState: ReplicationControllerState{
			TaskTemplate: TaskTemplate{
				DesiredState: TaskState{
					Manifest: ContainerManifest{
						Containers: []Container{
							Container{
								Image: "foo/bar",
							},
						},
					},
				},
				Labels: map[string]string{
					"name": "foo",
					"type": "production",
				},
			},
		},
	}

	taskControl.createReplica(controllerSpec)

	//expectedTask := Task{
	//	Labels:       controllerSpec.DesiredState.TaskTemplate.Labels,
	//	DesiredState: controllerSpec.DesiredState.TaskTemplate.DesiredState,
	//}
	// TODO: fix this so that it validates the body.
	fakeHandler.ValidateRequest(t, makeUrl("/tasks"), "POST", nil)
}

func TestHandleWatchResponseNotSet(t *testing.T) {
	body, _ := json.Marshal(makeTaskList(2))
	fakeHandler := util.FakeHandler{
		StatusCode:   200,
		ResponseBody: string(body),
	}
	testServer := httptest.NewTLSServer(&fakeHandler)
	client := Client{
		Host: testServer.URL,
	}

	fakeTaskControl := FakeTaskControl{}

	manager := MakeReplicationManager(nil, &client)
	manager.taskControl = &fakeTaskControl
	_, err := manager.handleWatchResponse(&etcd.Response{
		Action: "delete",
	})
	expectNoError(t, err)
}

func TestHandleWatchResponseNoNode(t *testing.T) {
	body, _ := json.Marshal(makeTaskList(2))
	fakeHandler := util.FakeHandler{
		StatusCode:   200,
		ResponseBody: string(body),
	}
	testServer := httptest.NewTLSServer(&fakeHandler)
	client := Client{
		Host: testServer.URL,
	}

	fakeTaskControl := FakeTaskControl{}

	manager := MakeReplicationManager(nil, &client)
	manager.taskControl = &fakeTaskControl
	_, err := manager.handleWatchResponse(&etcd.Response{
		Action: "set",
	})
	if err == nil {
		t.Error("Unexpected non-error")
	}
}

func TestHandleWatchResponseBadData(t *testing.T) {
	body, _ := json.Marshal(makeTaskList(2))
	fakeHandler := util.FakeHandler{
		StatusCode:   200,
		ResponseBody: string(body),
	}
	testServer := httptest.NewTLSServer(&fakeHandler)
	client := Client{
		Host: testServer.URL,
	}

	fakeTaskControl := FakeTaskControl{}

	manager := MakeReplicationManager(nil, &client)
	manager.taskControl = &fakeTaskControl
	_, err := manager.handleWatchResponse(&etcd.Response{
		Action: "set",
		Node: &etcd.Node{
			Value: "foobar",
		},
	})
	if err == nil {
		t.Error("Unexpected non-error")
	}
}

func TestHandleWatchResponse(t *testing.T) {
	body, _ := json.Marshal(makeTaskList(2))
	fakeHandler := util.FakeHandler{
		StatusCode:   200,
		ResponseBody: string(body),
	}
	testServer := httptest.NewTLSServer(&fakeHandler)
	client := Client{
		Host: testServer.URL,
	}

	fakeTaskControl := FakeTaskControl{}

	manager := MakeReplicationManager(nil, &client)
	manager.taskControl = &fakeTaskControl

	controller := makeReplicationController(2)

	data, err := json.Marshal(controller)
	expectNoError(t, err)
	controllerOut, err := manager.handleWatchResponse(&etcd.Response{
		Action: "set",
		Node: &etcd.Node{
			Value: string(data),
		},
	})
	if err != nil {
		t.Errorf("Unexpected error: %#v", err)
	}
	if !reflect.DeepEqual(controller, *controllerOut) {
		t.Errorf("Unexpected mismatch.  Expected %#v, Saw: %#v", controller, controllerOut)
	}
}
