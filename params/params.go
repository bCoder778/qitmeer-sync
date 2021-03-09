package params

type Params struct {
	BlockReward uint64
	GenesisUTXO map[string]uint64
}

var Qitmeer9Params = Params{
	BlockReward: 120 * 1e8,
	GenesisUTXO: map[string]uint64{
		"MEER": 6524293004366634,
	},
}

var Qitmeer10Params = Params{
	BlockReward: 120 * 1e8,
	GenesisUTXO: map[string]uint64{
		"MEER": 0,
		"QIT":  60000 * 1e8,
		"TER":  60000 * 1e8,
		"MET":  60000 * 1e8,
	},
}
