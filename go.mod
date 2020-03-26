module github.com/chartmuseum/storage

go 1.14

replace (
	github.com/NetEase-Object-Storage/nos-golang-sdk => github.com/karuppiah7890/nos-golang-sdk v0.0.0-20191116042345-0792ba35abcc
	go.etcd.io/etcd => github.com/eddycjy/etcd v0.5.0-alpha.5.0.20200218102753-4258cdd2efdf
)

require (
	cloud.google.com/go/storage v1.6.0
	github.com/Azure/azure-sdk-for-go v40.5.0+incompatible
	github.com/Azure/go-autorest/autorest v0.10.0 // indirect
	github.com/Azure/go-autorest/autorest/to v0.3.0 // indirect
	github.com/NetEase-Object-Storage/nos-golang-sdk v0.0.0-00010101000000-000000000000
	github.com/aliyun/aliyun-oss-go-sdk v2.0.7+incompatible
	github.com/aws/aws-sdk-go v1.29.32
	github.com/baidubce/bce-sdk-go v0.9.7
	github.com/baiyubin/aliyun-sts-go-sdk v0.0.0-20180326062324-cfa1a18b161f // indirect
	github.com/dnaeon/go-vcr v1.0.1 // indirect
	github.com/gophercloud/gophercloud v0.9.0
	github.com/oracle/oci-go-sdk v17.1.0+incompatible
	github.com/satori/go.uuid v1.2.0 // indirect
	github.com/stretchr/testify v1.5.1
	github.com/tencentyun/cos-go-sdk-v5 v0.0.0-20200311031902-5eb2ce1f587d
	go.etcd.io/etcd v0.0.0-00010101000000-000000000000
	golang.org/x/net v0.0.0-20200324143707-d3edc9973b7e
	google.golang.org/api v0.20.0
)
