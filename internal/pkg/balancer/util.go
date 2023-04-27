package balancer

import (
	"errors"
	"time"
)

var (
	timeout           time.Duration = 10 * time.Second
	ErrNoAliveServers               = errors.New("no alive replicas")
)
