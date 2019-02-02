# onlineConfig
Now, only support etcd...

## Install
`go get github.com/zhangxichun/onlineConfig`

## Usage
```golang
func main() {
    onlineConfig.NewEtcdWatch(onlineConfig.Config{
        Endpoints:   []string{"127.0.0.1:12379", "127.0.0.1:22379", "127.0.0.1:32379"}, // etcd cluster endpoints
        Root:        "app_configs", // optional, default is `app_configs`
        ServiceName: "event-broker", // your service name, in order to differentiate with other service
        Key:         "streams/order-cloud/services", // etcd key = Root + '/' + ServiceName + '/' + Key
        Onload:      reloadConfig, // a func type ( func([]byte) ), when get new config from etcd, this func will be excute
    })
}

type Test struct {
	A string
	B string
}

func reloadConfig(config []byte) { // config is value in etcd, in this case, the value in etcd is `{"A":"CCC","B":"DDD"}`
	test := &Test{}
	if err := json.Unmarshal(config, test); err != nil {
		logrus.Fatalf("Fail to unmarshal(%s).\n%s", config, err)
	}
	stdlog.Printf("Test: %s", test)
}
```