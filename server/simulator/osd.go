package simulator

import (
	"fmt"
	"sync"

	"github.com/mickeyzzc/onvif-go/v2/server/provider"
)

// osdStore is the simulator's in-memory OSD set: the create → set →
// delete loop for NVR OSD tests without hardware.
type osdStore struct {
	mu   sync.Mutex
	osds []provider.OSD
	next int
}

// OSDs lists the stored OSDs; a non-empty videoSourceConfigurationToken
// filters to that source.
func (s *osdStore) OSDs(videoSourceConfigurationToken string) []provider.OSD {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]provider.OSD, 0, len(s.osds))
	for _, o := range s.osds {
		if videoSourceConfigurationToken == "" || o.VideoSourceConfigurationToken == videoSourceConfigurationToken {
			out = append(out, o)
		}
	}

	return out
}

// UpsertOSD stores an OSD; an empty token assigns the next free one.
func (s *osdStore) UpsertOSD(osd provider.OSD) provider.OSD {
	s.mu.Lock()
	defer s.mu.Unlock()

	if osd.Type == "" {
		osd.Type = "Text"
	}
	if osd.Token == "" {
		s.next++
		osd.Token = fmt.Sprintf("osd_%d", s.next)
	}
	for i, o := range s.osds {
		if o.Token == osd.Token {
			s.osds[i] = osd

			return osd
		}
	}
	s.osds = append(s.osds, osd)

	return osd
}

// DeleteOSD removes an OSD by token; reports whether it existed.
func (s *osdStore) DeleteOSD(token string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, o := range s.osds {
		if o.Token == token {
			s.osds = append(s.osds[:i], s.osds[i+1:]...)

			return true
		}
	}

	return false
}
