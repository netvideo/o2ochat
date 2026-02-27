module github.com/netvideo/o2ochat

go 1.22

require (
	github.com/netvideo/cli v0.0.0
	github.com/netvideo/crypto v0.0.0
	github.com/netvideo/filetransfer v0.0.0
	github.com/netvideo/identity v0.0.0
	github.com/netvideo/media v0.0.0
	github.com/netvideo/signaling v0.0.0
	github.com/netvideo/storage v0.0.0
	github.com/netvideo/transport v0.0.0
	github.com/netvideo/ui v0.0.0
)

replace (
	github.com/netvideo/cli => ./cli
	github.com/netvideo/crypto => ./crypto
	github.com/netvideo/filetransfer => ./filetransfer
	github.com/netvideo/identity => ./identity
	github.com/netvideo/media => ./media
	github.com/netvideo/signaling => ./signaling
	github.com/netvideo/storage => ./storage
	github.com/netvideo/transport => ./transport
	github.com/netvideo/ui => ./ui
)
