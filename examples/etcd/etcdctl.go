package main

import (
	"context"
	"log"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func main() {
	var (
		err error
		key = "name"
		val = "pedro"
	)

	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: time.Second * 2,
		Username:    "",
		Password:    "",
	})
	if err != nil {
		log.Fatalf("%s", err)
	}

	ctx := context.Background()
	_, err = client.Put(ctx, key, val)
	if err != nil {
		log.Fatalf("put err: %s", err)
	}

	resp, err := client.Get(ctx, key)
	if err != nil {
		log.Fatalf("get err: %s", err)
	}
	log.Println(resp.Count, resp.Kvs)

	_, err = client.Delete(ctx, key)
	if err != nil {
		log.Fatalf("put err: %s", err)
	}

	resp, err = client.Get(ctx, key)
	if err != nil {
		log.Fatalf("get err: %s", err)
	}
	log.Println(resp.Count, resp.Kvs)
}
