package entity

import "time"

type WorkNode struct {
	ID         int64
	HostName   string
	Port       string
	NodeType   int32
	LaunchDate time.Time
	Created    time.Time
	Modified   time.Time
}
