package storage

import (

	"strings"
	"context"
	"fmt"
	"sync"
	"time"
	pathutil "path"

	"github.com/coreos/etcd/clientv3"

	"github.com/coreos/etcd/pkg/transport"
)

const DefaultPrefix = "/chart_backend_bucket"

var (
	DefileDialTimeOut ="5s"
	ErrNotExistEndpoints = fmt.Errorf("endpoints cannot connect !")
	ErrNotExist = fmt.Errorf("not exist!")
)

type etcdOpts struct {
	endpoints string
	cafile , certfile ,keyfile 	string
	dialtimeout time.Duration
}

type etcdStorage struct {
	c *clientv3.Client
	opts *etcdOpts

	base string
	bucket string
	ctx context.Context
	mu sync.RWMutex
}

//connection cut off
func isServerErr(err error) bool{
	if err != nil {
		if err == context.Canceled {
			return false
		} else if err == context.DeadlineExceeded {
			return false
		} else {
			return true
		}
	}
	return false
}

//prapare {basepath} dir
func (e *etcdStorage) probe() error{
	var (
		err error
	)
	ctx,cancel := context.WithCancel(e.ctx)
	_, err=e.c.Put(ctx,e.base,"")
	if err != nil && isServerErr(err) {
		return err
	}
	cancel()
	return nil
}

//
func (e *etcdStorage) ListObjects(prefix string) ([]Object, error) {
	var (
		objs []Object
	)
	ctx,cancel := context.WithTimeout(e.ctx,e.opts.dialtimeout)
	newpath := pathutil.Join(e.base, prefix)
	resps, err:=e.c.Get(ctx,newpath,clientv3.WithPrefix())
	cancel()
	if err!=nil {
		return nil ,err
	}
	for _,kv :=range resps.Kvs {
		if kv.Value != nil {
			path := removePrefixFromObjectPath(newpath, string(kv.Key))
			if objectPathIsInvalid(path) {
				continue
			}
			objs =append(objs, Object{
				Path:path,
				Content:kv.Value,
				LastModified: time.Unix(kv.ModRevision,0),
			})
		}
	}
	return objs,nil

}

func (e *etcdStorage) GetObject(path string) (Object, error){
	ctx,cancel := context.WithTimeout(e.ctx,e.opts.dialtimeout)
	newpath := pathutil.Join(e.base, path)
	resps, err:=e.c.Get(ctx,newpath)
	cancel()
	if err!=nil {
		return Object{} ,err
	}
	if len(resps.Kvs)!=1  || resps.Kvs[0].Value==nil {
		return Object{} ,ErrNotExist
	}
	return Object{
		Path:path,
		Content:resps.Kvs[0].Value,
		LastModified:time.Unix(resps.Kvs[0].ModRevision,0),
	},nil
}

func (e *etcdStorage) PutObject(path string, content []byte) error{
	ctx,cancel := context.WithTimeout(e.ctx,e.opts.dialtimeout)
	newpath := pathutil.Join(e.base, path)
	_, err:=e.c.Put(ctx,newpath,string(content))
	cancel()
	if err!=nil {
		return err
	}else{
		return nil
	}
}

func (e *etcdStorage) DeleteObject(path string) error{
	ctx,cancel := context.WithTimeout(e.ctx,e.opts.dialtimeout)
	newpath := pathutil.Join(e.base, path)
	_, err:=e.c.Delete(ctx,newpath)
	cancel()
	if err!=nil {
		return err
	}else{
		return nil
	}
}

func parseConf(endpoints string, cafile,certfile,keyfile string,dialtime time.Duration) clientv3.Config {
	var (
		es []string
	)
	if endpoints == ""{
		panic(ErrNotExistEndpoints)
	}
	es = strings.Split(endpoints,",")
	tlsInfo := transport.TLSInfo{
		CertFile:      certfile,
		KeyFile:       keyfile,
		CAFile: cafile,
	}
	tlsConfig, err := tlsInfo.ClientConfig()
	if err != nil {
		panic(err)
	}
	return  clientv3.Config{
		Endpoints:   es,
		DialTimeout: dialtime,
		TLS:         tlsConfig,
	}
}

func NewEtcdCSBackend(endpoints string, cafile,certfile,keyfile string,prefix string)  Backend {
	var (
		basepath string
	)
	DialTimeOut ,_ := time.ParseDuration(DefileDialTimeOut)
	cli, err := clientv3.New(parseConf(endpoints , cafile,certfile,keyfile ,DialTimeOut))
	if err != nil {
		panic(err)
	}

	if prefix == ""{
		basepath = DefaultPrefix
	}else {
		basepath = strings.TrimSuffix(prefix, "/")
		if basepath != "" && !strings.HasPrefix(basepath, "/") {
			basepath = "/" + basepath
		}
	}

	e:= &etcdStorage{
		c : cli,
		base:basepath,
		opts: &etcdOpts{
			endpoints:endpoints,
			cafile:cafile,
			dialtimeout:DialTimeOut,
			certfile:certfile,
			keyfile:keyfile,
		},
		ctx:context.Background(),
	}
	if err := e.probe();err!=nil{
		panic(err)
	}
	return e


}