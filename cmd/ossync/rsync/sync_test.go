package rsync

import (
	"fmt"
	"os"
	"testing"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/pubgo/xerror"
)

func TestName(t *testing.T) {
	client, err := oss.New(
		os.Getenv("oss_endpoint"),
		os.Getenv("oss_ak"),
		os.Getenv("oss_sk"),
	)
	xerror.Panic(err)

	continuationToken := ""
	for i := 2; i > 0; i-- {
		kk := xerror.PanicErr(client.Bucket("kooksee")).(*oss.Bucket)
		resp := xerror.PanicErr(kk.ListObjectsV2(oss.Prefix(syncPrefix), oss.ContinuationToken(continuationToken))).(oss.ListObjectsResultV2)
		continuationToken = resp.NextContinuationToken

		fmt.Println(resp.Prefix)
		fmt.Println(resp.XMLName)
		fmt.Println(resp.MaxKeys)
		fmt.Println(resp.MaxKeys)
		fmt.Println(resp.Delimiter)
		fmt.Println(resp.IsTruncated)
		fmt.Println(resp.CommonPrefixes)
		for _, k := range resp.Objects {
			fmt.Printf("%#v\n", k)
		}
	}
}

//携带文件的Content-MD5值
