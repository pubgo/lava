package oss

import (
	"fmt"
	"os"
	"testing"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
)

func TestName(t *testing.T) {
	client, err := oss.New(
		os.Getenv("oss_endpoint"),
		os.Getenv("oss_ak"),
		os.Getenv("oss_sk"))
	xerror.Panic(err)

	lsRes, err := client.ListBuckets()
	xerror.Panic(err)

	for _, bucket := range lsRes.Buckets {
		fmt.Println("Buckets:", bucket.Name)
	}
}

func TestName1(t *testing.T) {
	xlog.Debug("ss", xlog.Any("err", xerror.New("ssss")))
}
