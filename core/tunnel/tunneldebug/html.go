package tunneldebug

import _ "embed"

//go:embed gateway.html
var gatewayDashboardHTML string

//go:embed agent.html
var agentDashboardHTML string

//go:embed empty.html
var emptyDashboardHTML string

// getGatewayDashboardHTML 返回 Gateway 仪表盘 HTML
func getGatewayDashboardHTML() string {
	return gatewayDashboardHTML
}

// getAgentDashboardHTML 返回 Agent 仪表盘 HTML
func getAgentDashboardHTML() string {
	return agentDashboardHTML
}

// getEmptyDashboardHTML 返回空状态页面
func getEmptyDashboardHTML() string {
	return emptyDashboardHTML
}
