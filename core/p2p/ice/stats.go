package ice

import (
	pionice "github.com/pion/ice/v4"
)

// LivePairStats 从仍存活的 ICE Agent 读取选路与 RTT（供 debug / metrics）。
type LivePairStats struct {
	Pair      CandidatePairInfo
	RTTMs     float64
	PairState string
	Nominated bool
}

// LivePairStatsFromAgent 读取当前选中候选对的实时统计；agent 为 nil 或无选路时返回 false。
func LivePairStatsFromAgent(agent *pionice.Agent) (LivePairStats, bool) {
	if agent == nil {
		return LivePairStats{}, false
	}
	pair, ok := SelectedPairInfo(agent)
	if !ok {
		return LivePairStats{}, false
	}
	st := LivePairStats{Pair: pair}
	if ps, ok := agent.GetSelectedCandidatePairStats(); ok {
		st.RTTMs = ps.CurrentRoundTripTime * 1000
		st.PairState = ps.State.String()
		st.Nominated = ps.Nominated
	}
	return st, true
}
