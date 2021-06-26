package routerlist

import "github.com/11th-ndn-hackathon/ndn-fch/model"

var yoursunnyRouters = []model.Router{
	{
		ID:        "yoursunny-quic-lax",
		Position:  model.LonLat{-118.2437, 34.0522},
		Prefix:    "/yoursunny/_/lax",
		Host:      "lax.quic.g.ndn.today",
		IPv4:      true,
		HTTP3Port: 6367,
	}, {
		ID:        "yoursunny-quic-dal",
		Position:  model.LonLat{-96.7970, 32.7767},
		Prefix:    "/yoursunny/_/dal",
		Host:      "dal.quic.g.ndn.today",
		IPv4:      true,
		IPv6:      true,
		HTTP3Port: 6367,
	}, {
		ID:       "yoursunny-udp-yul",
		Position: model.LonLat{-73.8779, 45.3151},
		Prefix:   "/yoursunny/_/yul",
		Host:     "yul.udp.g.ndn.today",
		IPv6:     true,
		UDPPort:  6363,
	}, {
		ID:        "yoursunny-quic-lil",
		Position:  model.LonLat{3.1778, 50.6927},
		Prefix:    "/yoursunny/_/lil",
		Host:      "lil.quic.g.ndn.today",
		IPv4:      true,
		IPv6:      true,
		HTTP3Port: 6367,
	}, {
		ID:            "yoursunny-ws-muc",
		Position:      model.LonLat{11.5820, 48.1351},
		Prefix:        "/yoursunny/_/muc",
		Host:          "muc.ws.g.ndn.today",
		IPv6:          true,
		WebSocketPort: 443,
	}, {
		ID:        "yoursunny-quic-waw",
		Position:  model.LonLat{21.0122, 52.2297},
		Prefix:    "/yoursunny/_/waw",
		Host:      "waw.quic.g.ndn.today",
		IPv6:      true,
		HTTP3Port: 10207,
	}, {
		ID:        "yoursunny-quic-nrt",
		Position:  model.LonLat{139.7690, 35.6804},
		Prefix:    "/yoursunny/_/nrt",
		Host:      "nrt.quic.g.ndn.today",
		IPv4:      true,
		IPv6:      true,
		HTTP3Port: 6367,
	},
}
