package driverloc

import (
	"time"
)

func (s *ServiceImpl) SetTimeNowFn(fn func() time.Time) { s.timeNowFn = fn }
