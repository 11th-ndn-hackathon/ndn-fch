package model_test

import (
	"testing"

	"github.com/11th-ndn-hackathon/ndn-fch/model"
	"github.com/stretchr/testify/assert"
)

func TestQuery(t *testing.T) {
	assert := assert.New(t)

	list := model.ParseQueries("k=3&cap=udp&k=2&cap=http3&ipv4=1&ipv6=1&lon=121.4737&lat=31.2304")
	assert.Len(list, 2)
	for i, q := range list {
		if i == 0 {
			assert.Equal(3, q.Count)
			assert.Equal(model.TransportUDP, q.Transport)
		} else {
			assert.Equal(2, q.Count)
			assert.Equal(model.TransportH3, q.Transport)
		}
		assert.True(q.IPv4)
		assert.True(q.IPv6)
		assert.InDelta(121.4737, q.Position[0], 0.0001)
		assert.InDelta(31.2304, q.Position[1], 0.0001)
	}
}
