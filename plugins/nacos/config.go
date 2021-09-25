package nacos

import (
	"fmt"

	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

var Name = "nacos"

func init() {
	cc := constant.ClientConfig{
		Endpoint:    "acm.aliyun.com:8080",
		NamespaceId: "e525eafa-f7d7-4029-83d9-008937f9d468",
		RegionId:    "cn-shanghai",
		AccessKey:   "LTAI4G8KxxxxxxxxxxxxxbwZLBr",
		SecretKey:   "n5jTL9YxxxxxxxxxxxxaxmPLZV9",
		OpenKMS:     true,
		TimeoutMs:   5000,
	}

	// a more graceful way to create config client
	client, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig: &cc,
		},
	)

	if err != nil {
		panic(err)
	}

	// to enable encrypt/decrypt, DataId should be start with "cipher-"
	_, err = client.PublishConfig(vo.ConfigParam{
		DataId:  "cipher-dataId-1",
		Group:   "test-group",
		Content: "hello world!",
	})

	if err != nil {
		fmt.Printf("PublishConfig err: %v\n", err)
	}

	//get config
	content, err := client.GetConfig(vo.ConfigParam{
		DataId: "cipher-dataId-3",
		Group:  "test-group",
	})
	fmt.Printf("GetConfig, config: %s, error: %v\n", content, err)

	// DataId is not start with "cipher-", content will not be encrypted.
	_, err = client.PublishConfig(vo.ConfigParam{
		DataId:  "dataId-1",
		Group:   "test-group",
		Content: "hello world!",
	})

	if err != nil {
		fmt.Printf("PublishConfig err: %v\n", err)
	}

	//get config
	content, err = client.GetConfig(vo.ConfigParam{
		DataId: "dataId-1",
		Group:  "test-group",
	})
	fmt.Printf("GetConfig, config: %s, error: %v\n", content, err)
}
