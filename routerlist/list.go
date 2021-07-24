package routerlist

import "github.com/11th-ndn-hackathon/ndn-fch/model"

// List returns a list of known routers.
func List() (routers []model.Router) {
	testbedRoutersLock.RLock()
	defer testbedRoutersLock.RUnlock()
	routers = append(routers, testbedRouters...)
	routers = append(routers, ndn6Routers...)
	return routers
}

// Load initializes the list.
func Load() {
	loadNDN6Topo()
	updateTestbedRouters()
}
