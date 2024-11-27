package run

import "sync"

type UpstreamManager struct {
	dialers *sync.Pool
}

func NewUpstreamManager() {

}
