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
package kubectl

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/pedrogao/simplek8s/pkg/api"
	"github.com/pedrogao/simplek8s/pkg/util"
)

// TODO: This doesn't reduce typing enough to make it worth the less readable errors. Remove.
func expectNoError(t *testing.T, err error) {
	if err != nil {
		t.Errorf("Unexpected error: %#v", err)
	}
}

type Action struct {
	action string
	value  interface{}
}

type FakeKubeClient struct {
	actions []Action
	tasks   TaskList
	ctrl    ReplicationController
}

func (client *FakeKubeClient) ListTasks(labelQuery map[string]string) (TaskList, error) {
	client.actions = append(client.actions, Action{action: "list-tasks"})
	return client.tasks, nil
}

func (client *FakeKubeClient) GetTask(name string) (Task, error) {
	client.actions = append(client.actions, Action{action: "get-task", value: name})
	return Task{}, nil
}

func (client *FakeKubeClient) DeleteTask(name string) error {
	client.actions = append(client.actions, Action{action: "delete-task", value: name})
	return nil
}

func (client *FakeKubeClient) CreateTask(task Task) (Task, error) {
	client.actions = append(client.actions, Action{action: "create-task"})
	return Task{}, nil
}

func (client *FakeKubeClient) UpdateTask(task Task) (Task, error) {
	client.actions = append(client.actions, Action{action: "update-task", value: task.ID})
	return Task{}, nil
}

func (client *FakeKubeClient) GetReplicationController(name string) (ReplicationController, error) {
	client.actions = append(client.actions, Action{action: "get-controller", value: name})
	return client.ctrl, nil
}

func (client *FakeKubeClient) CreateReplicationController(controller ReplicationController) (ReplicationController, error) {
	client.actions = append(client.actions, Action{action: "create-controller", value: controller})
	return ReplicationController{}, nil
}

func (client *FakeKubeClient) UpdateReplicationController(controller ReplicationController) (ReplicationController, error) {
	client.actions = append(client.actions, Action{action: "update-controller", value: controller})
	return ReplicationController{}, nil
}

func (client *FakeKubeClient) DeleteReplicationController(controller string) error {
	client.actions = append(client.actions, Action{action: "delete-controller", value: controller})
	return nil
}

func (client *FakeKubeClient) GetService(name string) (Service, error) {
	client.actions = append(client.actions, Action{action: "get-controller", value: name})
	return Service{}, nil
}

func (client *FakeKubeClient) CreateService(controller Service) (Service, error) {
	client.actions = append(client.actions, Action{action: "create-service", value: controller})
	return Service{}, nil
}

func (client *FakeKubeClient) UpdateService(controller Service) (Service, error) {
	client.actions = append(client.actions, Action{action: "update-service", value: controller})
	return Service{}, nil
}

func (client *FakeKubeClient) DeleteService(controller string) error {
	client.actions = append(client.actions, Action{action: "delete-service", value: controller})
	return nil
}

func validateAction(expectedAction, actualAction Action, t *testing.T) {
	if expectedAction != actualAction {
		t.Errorf("Unexpected action: %#v, expected: %#v", actualAction, expectedAction)
	}
}

func TestUpdateWithTasks(t *testing.T) {
	client := FakeKubeClient{
		tasks: TaskList{
			Items: []Task{
				Task{JSONBase: JSONBase{ID: "task-1"}},
				Task{JSONBase: JSONBase{ID: "task-2"}},
			},
		},
	}
	Update("foo", &client, 0)
	if len(client.actions) != 4 {
		t.Errorf("Unexpected action list %#v", client.actions)
	}
	validateAction(Action{action: "get-controller", value: "foo"}, client.actions[0], t)
	validateAction(Action{action: "list-tasks"}, client.actions[1], t)
	validateAction(Action{action: "update-task", value: "task-1"}, client.actions[2], t)
	validateAction(Action{action: "update-task", value: "task-2"}, client.actions[3], t)
}

func TestUpdateNoTasks(t *testing.T) {
	client := FakeKubeClient{}
	Update("foo", &client, 0)
	if len(client.actions) != 2 {
		t.Errorf("Unexpected action list %#v", client.actions)
	}
	validateAction(Action{action: "get-controller", value: "foo"}, client.actions[0], t)
	validateAction(Action{action: "list-tasks"}, client.actions[1], t)
}

func TestDoRequest(t *testing.T) {
	expectedBody := `{ "items": []}`
	fakeHandler := util.FakeHandler{
		StatusCode:   200,
		ResponseBody: expectedBody,
	}
	testServer := httptest.NewTLSServer(&fakeHandler)
	request, _ := http.NewRequest("GET", testServer.URL+"/foo/bar", nil)
	body, err := DoRequest(request, "user", "pass")
	if request.Header["Authorization"] == nil {
		t.Errorf("Request is missing authorization header: %#v", *request)
	}
	if err != nil {
		t.Error("Unexpected error")
	}
	if body != expectedBody {
		t.Errorf("Expected body: '%s', saw: '%s'", expectedBody, body)
	}
	fakeHandler.ValidateRequest(t, "/foo/bar", "GET", &fakeHandler.ResponseBody)
}

func TestRunController(t *testing.T) {
	fakeClient := FakeKubeClient{}
	name := "name"
	image := "foo/bar"
	replicas := 3
	RunController(image, name, replicas, &fakeClient, "8080:80", -1)
	if len(fakeClient.actions) != 1 || fakeClient.actions[0].action != "create-controller" {
		t.Errorf("Unexpected actions: %#v", fakeClient.actions)
	}
	controller := fakeClient.actions[0].value.(ReplicationController)
	if controller.ID != name ||
		controller.DesiredState.Replicas != replicas ||
		controller.DesiredState.TaskTemplate.DesiredState.Manifest.Containers[0].Image != image {
		t.Errorf("Unexpected controller: %#v", controller)
	}
}

func TestRunControllerWithService(t *testing.T) {
	fakeClient := FakeKubeClient{}
	name := "name"
	image := "foo/bar"
	replicas := 3
	RunController(image, name, replicas, &fakeClient, "", 8000)
	if len(fakeClient.actions) != 2 ||
		fakeClient.actions[0].action != "create-controller" ||
		fakeClient.actions[1].action != "create-service" {
		t.Errorf("Unexpected actions: %#v", fakeClient.actions)
	}
	controller := fakeClient.actions[0].value.(ReplicationController)
	if controller.ID != name ||
		controller.DesiredState.Replicas != replicas ||
		controller.DesiredState.TaskTemplate.DesiredState.Manifest.Containers[0].Image != image {
		t.Errorf("Unexpected controller: %#v", controller)
	}
}

func TestStopController(t *testing.T) {
	fakeClient := FakeKubeClient{}
	name := "name"
	StopController(name, &fakeClient)
	if len(fakeClient.actions) != 2 {
		t.Errorf("Unexpected actions: %#v", fakeClient.actions)
	}
	if fakeClient.actions[0].action != "get-controller" ||
		fakeClient.actions[0].value.(string) != name {
		t.Errorf("Unexpected action: %#v", fakeClient.actions[0])
	}
	controller := fakeClient.actions[1].value.(ReplicationController)
	if fakeClient.actions[1].action != "update-controller" ||
		controller.DesiredState.Replicas != 0 {
		t.Errorf("Unexpected action: %#v", fakeClient.actions[1])
	}
}

func TestKubectlDeleteController(t *testing.T) {
	fakeClient := FakeKubeClient{}
	name := "name"
	err := DeleteController(name, &fakeClient)
	expectNoError(t, err)
	if len(fakeClient.actions) != 2 {
		t.Errorf("Unexpected actions: %#v", fakeClient.actions)
	}
	if fakeClient.actions[0].action != "get-controller" ||
		fakeClient.actions[0].value.(string) != name {
		t.Errorf("Unexpected action: %#v", fakeClient.actions[0])
	}
	if fakeClient.actions[1].action != "delete-controller" ||
		fakeClient.actions[1].value.(string) != name {
		t.Errorf("Unexpected action: %#v", fakeClient.actions[1])
	}
}

func TestKubectlDeleteControllerWithReplicas(t *testing.T) {
	fakeClient := FakeKubeClient{
		ctrl: ReplicationController{
			DesiredState: ReplicationControllerState{
				Replicas: 2,
			},
		},
	}
	name := "name"
	err := DeleteController(name, &fakeClient)
	if len(fakeClient.actions) != 1 {
		t.Errorf("Unexpected actions: %#v", fakeClient.actions)
	}
	if fakeClient.actions[0].action != "get-controller" ||
		fakeClient.actions[0].value.(string) != name {
		t.Errorf("Unexpected action: %#v", fakeClient.actions[0])
	}
	if err == nil {
		t.Errorf("Unexpected non-error.")
	}
}

func TestRequestWithBodyNoSuchFile(t *testing.T) {
	request, err := RequestWithBody("non/existent/file.json", "http://www.google.com", "GET")
	if request != nil {
		t.Error("Unexpected non-nil result")
	}
	if err == nil {
		t.Error("Unexpected non-error")
	}
}

func TestRequestWithBody(t *testing.T) {
	file, err := ioutil.TempFile("", "foo")
	expectNoError(t, err)
	data, err := json.Marshal(Task{JSONBase: JSONBase{ID: "foo"}})
	expectNoError(t, err)
	_, err = file.Write(data)
	expectNoError(t, err)
	request, err := RequestWithBody(file.Name(), "http://www.google.com", "GET")
	if request == nil {
		t.Error("Unexpected nil result")
	}
	if err != nil {
		t.Errorf("Unexpected error: %#v", err)
	}
	dataOut, err := ioutil.ReadAll(request.Body)
	expectNoError(t, err)
	if string(data) != string(dataOut) {
		t.Errorf("Mismatched data. Expected %s, got %s", data, dataOut)
	}
}

func validatePort(t *testing.T, p Port, external int, internal int) {
	if p.HostPort != external || p.ContainerPort != internal {
		t.Errorf("Unexpected port: %#v != (%d, %d)", p, external, internal)
	}
}

func TestMakePorts(t *testing.T) {
	ports := makePorts("8080:80,8081:8081,443:444")
	if len(ports) != 3 {
		t.Errorf("Unexpected ports: %#v", ports)
	}

	validatePort(t, ports[0], 8080, 80)
	validatePort(t, ports[1], 8081, 8081)
	validatePort(t, ports[2], 443, 444)
}
