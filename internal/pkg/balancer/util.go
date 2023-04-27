package balancer

import (
	"errors"
	"time"
)

var (
	timeout           time.Duration = 5 * time.Second
	ErrNoAliveServers               = errors.New("no alive replicas")
)
