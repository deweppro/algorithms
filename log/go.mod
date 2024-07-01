module go.osspkg.com/x/log

go 1.20

replace (
	go.osspkg.com/x/sync => ../sync
	go.osspkg.com/x/test => ../test
)

require (
	github.com/mailru/easyjson v0.7.7
	go.osspkg.com/x/sync v0.5.0
	go.osspkg.com/x/test v0.5.0
)

require github.com/josharian/intern v1.0.0 // indirect