package beatsmith

type MasteringSettings struct {
	EQ       string `json:"eq"`
	Comp     string `json:"compression"`
	Limiter  string `json:"limiter"`
}

type MasteringAssistant struct{}

func (m *MasteringAssistant) RecommendSettings(genre string) MasteringSettings {
	switch genre {
	case "Trap":
		return MasteringSettings{
			EQ:      "Boost lows at 50Hz, cut mud at 250Hz, boost highs at 10kHz.",
			Comp:    "Fast attack, medium release, 2:1 ratio for glue.",
			Limiter: "-0.1dB ceiling, aggressive lookahead.",
		}
	case "Lofi":
		return MasteringSettings{
			EQ:      "High cut at 15kHz, low cut at 40Hz, boost mids slightly.",
			Comp:    "Slow attack, slow release for pumping effect.",
			Limiter: "-1.0dB ceiling, soft clip.",
		}
	default:
		return MasteringSettings{
			EQ:      "Transparent curve.",
			Comp:    "Gentle bus compression.",
			Limiter: "-0.3dB ceiling.",
		}
	}
}
