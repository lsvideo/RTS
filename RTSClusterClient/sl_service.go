// sl_service
package main

//定义服务interface
type SL_Service interface {
	SLServiceStart()
	SLServiceStop()
}

var mapService = make(map[ServerType]SL_Service, 5)

type ServiceUnknown struct {
}

var unService ServiceUnknown

func init() {
	log.Println("init nill service")
	mapService[ServerTypeUnknown] = unService
}

func (un ServiceUnknown) SLServiceStart() {
	log.Errorf("%s:%s unknown service type!", APP_NAME, GetFuncName())
	serverexit <- 1
}

func (un ServiceUnknown) SLServiceStop() {

}
