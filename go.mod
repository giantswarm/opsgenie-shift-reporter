module github.com/giantswarm/oncall-shift-reporter

go 1.23

require (
	github.com/PagerDuty/go-pagerduty v1.8.0
	github.com/nlopes/slack v0.6.0
)

require (
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/gorilla/websocket v1.2.0 // indirect
	github.com/pkg/errors v0.8.1 // indirect
)

// Force-upgrade transitive deps flagged by the nancy security scan.
// go-pagerduty v1.8.0 (latest release) directly requires the vulnerable
// versions, so they survive module-graph pruning and cannot be bumped via a
// plain require (go mod tidy reverts unimported deps). These modules are not
// compiled into the binary; the replace only raises the version nancy scans.
replace golang.org/x/crypto => golang.org/x/crypto v0.57.0

replace golang.org/x/sys => golang.org/x/sys v0.48.0
