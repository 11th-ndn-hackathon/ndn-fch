package routerlist

var yoursunnyRouters = []Router{
	{
		ID:        "yoursunny-lax",
		Position:  LonLat{-118.2437, 34.0522},
		Prefix:    "/yoursunny/_/lax",
		Host:      "lax.quic.g.ndn.today",
		IPv4:      true,
		HTTP3Port: 6367,
	}, {
		ID:        "yoursunny-dal",
		Position:  LonLat{-96.7970, 32.7767},
		Prefix:    "/yoursunny/_/dal",
		Host:      "dal.quic.g.ndn.today",
		IPv4:      true,
		IPv6:      true,
		HTTP3Port: 6367,
	}, {
		ID:        "yoursunny-lil",
		Position:  LonLat{3.1778, 50.6927},
		Prefix:    "/yoursunny/_/lil",
		Host:      "lil.quic.g.ndn.today",
		IPv4:      true,
		IPv6:      true,
		HTTP3Port: 6367,
	}, {
		ID:        "yoursunny-waw",
		Position:  LonLat{21.0122, 52.2297},
		Prefix:    "/yoursunny/_/waw",
		Host:      "waw.quic.g.ndn.today",
		IPv6:      true,
		HTTP3Port: 10207,
	}, {
		ID:        "yoursunny-nrt",
		Position:  LonLat{139.7690, 35.6804},
		Prefix:    "/yoursunny/_/nrt",
		Host:      "nrt.quic.g.ndn.today",
		IPv4:      true,
		IPv6:      true,
		HTTP3Port: 6367,
	},
}
