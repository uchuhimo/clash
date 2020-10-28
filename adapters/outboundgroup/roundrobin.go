package outboundgroup

import (
	"context"
	"encoding/json"

	"github.com/Dreamacro/clash/adapters/outbound"
	"github.com/Dreamacro/clash/adapters/provider"
	"github.com/Dreamacro/clash/common/singledo"
	C "github.com/Dreamacro/clash/constant"
)

type RoundRobin struct {
	*outbound.Base
	single    *singledo.Single
	index     int
	providers []provider.ProxyProvider
}


func (rr *RoundRobin) DialContext(ctx context.Context, metadata *C.Metadata) (c C.Conn, err error) {
	defer func() {
		if err == nil {
			c.AppendToChains(rr)
		}
	}()

	proxy := rr.Unwrap(metadata)

	c, err = proxy.DialContext(ctx, metadata)
	return
}

func (rr *RoundRobin) DialUDP(metadata *C.Metadata) (pc C.PacketConn, err error) {
	defer func() {
		if err == nil {
			pc.AppendToChains(rr)
		}
	}()

	proxy := rr.Unwrap(metadata)

	return proxy.DialUDP(metadata)
}

func (rr *RoundRobin) SupportUDP() bool {
	return true
}

func (rr *RoundRobin) Unwrap(metadata *C.Metadata) C.Proxy {
	proxies := rr.proxies()
	for i := 0; i < len(proxies); i++ {
		rr.index = (rr.index + 1) % len(proxies)
		proxy := proxies[rr.index]
		if proxy.Alive() {
			return proxy
		}
	}

	return proxies[0]
}

func (rr *RoundRobin) proxies() []C.Proxy {
	elm, _, _ := rr.single.Do(func() (interface{}, error) {
		return getProvidersProxies(rr.providers), nil
	})

	return elm.([]C.Proxy)
}

func (rr *RoundRobin) MarshalJSON() ([]byte, error) {
	var all []string
	for _, proxy := range rr.proxies() {
		all = append(all, proxy.Name())
	}
	return json.Marshal(map[string]interface{}{
		"type": rr.Type().String(),
		"all":  all,
	})
}

func NewRoundRobin(name string, providers []provider.ProxyProvider) *RoundRobin {
	return &RoundRobin{
		Base:      outbound.NewBase(name, "", C.LoadBalance, false),
		single:    singledo.NewSingle(defaultGetProxiesDuration),
		index:  0,
		providers: providers,
	}
}
