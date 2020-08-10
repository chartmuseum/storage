package storage

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

var  (
	cafile ="./_test/ca.pem"
	certfile ="./_test/cert.pem"
	keyfile ="./_test/key.pem"
	endpoints = "https://127.0.0.1:2379"
)

type CsEtcdSuite struct {
	suite.Suite
	etcd Backend
}

func (c *CsEtcdSuite) SetupSuite() {
	// TODO: renable, see https://github.com/chartmuseum/storage/issues/6
	c.T().Skip()
	c.etcd=NewEtcdCSBackend(endpoints,cafile,certfile,keyfile,"")
	fmt.Println("init success")
}


func (c *CsEtcdSuite) TestListObjecs() {
	c.etcd.PutObject("listtest1",[]byte("puttestetcd"))

	c.etcd.PutObject("listtest2",[]byte("puttestetcd"))

	if objs,err:=c.etcd.ListObjects("listtest");err!=nil{
		c.Fail("etcd list objects err:%s",err)
	}else{
		fmt.Println(objs)
	}

}

func (c *CsEtcdSuite) TestGetObject(){

	c.etcd.PutObject("gettest",[]byte("testdate"))

	if obj,err:=c.etcd.GetObject("gettest");err!=nil{
		c.Fail("etcd get objects err:%s",err)
	}else{
		fmt.Println(obj)
	}

}

func (c *CsEtcdSuite) TestDeleteObject(){

	c.etcd.PutObject("deletetest",[]byte("testdate"))

	if err:=c.etcd.DeleteObject("deletetest");err!=nil{
		c.Fail("etcd delete objects err:%s",err)
	}

}

func TestEtcdCSBackend(t *testing.T) {

	suite.Run(t, new(CsEtcdSuite))
}
