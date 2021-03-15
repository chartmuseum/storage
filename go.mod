module github.com/chartmuseum/storage

go 1.15

replace (
	github.com/NetEase-Object-Storage/nos-golang-sdk => github.com/karuppiah7890/nos-golang-sdk v0.0.0-20191116042345-0792ba35abcc
	go.etcd.io/etcd => github.com/eddycjy/etcd v0.5.0-alpha.5.0.20200218102753-4258cdd2efdf
	google.golang.org/grpc => google.golang.org/grpc v1.29.1
)

require (
	cloud.google.com/go/storage v1.13.0
	github.com/Azure/azure-sdk-for-go v48.2.2+incompatible
	github.com/Azure/go-autorest/autorest v0.11.11 // indirect
	github.com/Azure/go-autorest/autorest/to v0.4.0 // indirect
	github.com/NetEase-Object-Storage/nos-golang-sdk v0.0.0-00010101000000-000000000000
	github.com/aliyun/aliyun-oss-go-sdk v2.1.6+incompatible
	github.com/aws/aws-sdk-go v1.37.30
	github.com/baidubce/bce-sdk-go v0.9.49
	github.com/baiyubin/aliyun-sts-go-sdk v0.0.0-20180326062324-cfa1a18b161f // indirect
	github.com/dnaeon/go-vcr v1.1.0 // indirect
	github.com/gophercloud/gophercloud v0.16.0
	github.com/oracle/oci-go-sdk v24.3.0+incompatible
	github.com/satori/go.uuid v1.2.0 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/tencentyun/cos-go-sdk-v5 v0.7.21
	go.etcd.io/etcd v0.0.0-00010101000000-000000000000
	golang.org/x/net v0.0.0-20210119194325-5f4716e94777
	google.golang.org/api v0.41.0
)
