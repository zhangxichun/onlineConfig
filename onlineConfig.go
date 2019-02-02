package onlineConfig

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"go.etcd.io/etcd/clientv3"
)

const _root = "app_configs"

type Config struct {
	Endpoints   []string
	Root        string
	ServiceName string
	Key         string
	Onload      func([]byte)
}

var etcdClient *clientv3.Client

func NewEtcdWatch(config Config) error {
	if err := checkFields(config); err != nil {
		return err
	}

	root := _root
	if len(config.Root) > 0 {
		root = config.Root
	}

	key := []string{root, config.ServiceName, config.Key}
	config.Key = strings.Join(key, "/")
	logrus.Infof("Watching Key(%s)", config.Key)
	var err error
	etcdClient, err = clientv3.New(clientv3.Config{
		Endpoints:   config.Endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return err
	}
	return watchKey(config.Key, config.Onload)
}

func watchKey(key string, handle func([]byte)) error {
	timeout := 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	resp, err := etcdClient.Get(ctx, key)
	cancel()
	if err != nil {
		return err
	}
	if len(resp.Kvs) == 0 {
		return errors.New("KeyValue is empty.")
	}
	ev := resp.Kvs[0]
	handle(ev.Value)

	go watchEtcd(*etcdClient, key, handle)

	return nil
}

func watchEtcd(cli clientv3.Client, key string, handle func(value []byte)) {
	rch := cli.Watch(context.Background(), key)
	for wresp := range rch {
		logrus.Infof("Reload config...")
		if err := wresp.Err(); err != nil {
			logrus.Fatalf("Fail to watch [%s] from etcd.\n%s", key, err)
		}
		if len(wresp.Events) == 0 {
			logrus.Fatalf("Fail to watch [%s] from etcd.\nEvents is empty.", key)
		}
		for _, ev := range wresp.Events {
			switch ev.Type {
			case clientv3.EventTypePut:
				handle(ev.Kv.Value)
			case clientv3.EventTypeDelete:
				logrus.Fatalf("config has been deleted.")
			}

		}
	}
}

func checkFields(config Config) error {
	if len(config.Endpoints) == 0 {
		return errors.New("Endpoints can't be empty.")
	}
	if len(config.ServiceName) == 0 {
		return errors.New("ServiceName can't be empty.")
	}
	if len(config.Key) == 0 {
		return errors.New("Key can't be empty.")
	}
	if config.Onload == nil {
		return errors.New("Onload can't be empty.")
	}

	return nil
}
