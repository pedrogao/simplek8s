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
	"net/url"

	. "github.com/pedrogao/simplek8s/pkg/api"
	"github.com/pedrogao/simplek8s/pkg/apiserver"
)

// Implementation of RESTStorage for the api server.
type ControllerRegistryStorage struct {
	registry ControllerRegistry
}

func MakeControllerRegistryStorage(registry ControllerRegistry) apiserver.RESTStorage {
	return &ControllerRegistryStorage{
		registry: registry,
	}
}

func (storage *ControllerRegistryStorage) List(*url.URL) (interface{}, error) {
	var result ReplicationControllerList
	controllers, err := storage.registry.ListControllers()
	if err == nil {
		result = ReplicationControllerList{
			Items: controllers,
		}
	}
	return result, err
}

func (storage *ControllerRegistryStorage) Get(id string) (interface{}, error) {
	return storage.registry.GetController(id)
}

func (storage *ControllerRegistryStorage) Delete(id string) error {
	return storage.registry.DeleteController(id)
}

func (storage *ControllerRegistryStorage) Extract(body string) (interface{}, error) {
	result := ReplicationController{}
	err := json.Unmarshal([]byte(body), &result)
	return result, err
}

func (storage *ControllerRegistryStorage) Create(controller interface{}) error {
	return storage.registry.CreateController(controller.(ReplicationController))
}

func (storage *ControllerRegistryStorage) Update(controller interface{}) error {
	return storage.registry.UpdateController(controller.(ReplicationController))
}
