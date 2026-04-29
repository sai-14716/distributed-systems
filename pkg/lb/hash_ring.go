package lb

import (
	"fmt"
	"hash/crc32"
	"sort"
)

// hashRing is a basic consistent-hashing ring with virtual nodes.
// Owner = first node clockwise from key hash.
type hashRing struct {
	points []ringPoint
	nodes  []string
}

type ringPoint struct {
	h      uint32
	nodeID string
}

func newHashRing(peers []Peer, replicas int) (*hashRing, error) {
	if replicas < 1 {
		replicas = 1
	}
	ids := make([]string, 0, len(peers))
	seen := map[string]struct{}{}
	for _, p := range peers {
		if p.ID == "" {
			continue
		}
		if _, ok := seen[p.ID]; ok {
			continue
		}
		seen[p.ID] = struct{}{}
		ids = append(ids, p.ID)
	}
	sort.Strings(ids)
	if len(ids) == 0 {
		return nil, fmt.Errorf("hash ring requires at least one peer")
	}
	points := make([]ringPoint, 0, len(ids)*replicas)
	for _, id := range ids {
		for i := 0; i < replicas; i++ {
			k := fmt.Sprintf("%s#%d", id, i)
			points = append(points, ringPoint{
				h:      crc32.ChecksumIEEE([]byte(k)),
				nodeID: id,
			})
		}
	}
	sort.Slice(points, func(i, j int) bool {
		if points[i].h == points[j].h {
			return points[i].nodeID < points[j].nodeID
		}
		return points[i].h < points[j].h
	})
	return &hashRing{points: points, nodes: ids}, nil
}

func (r *hashRing) owner(key string) string {
	if r == nil || len(r.points) == 0 {
		return ""
	}
	h := crc32.ChecksumIEEE([]byte(key))
	idx := sort.Search(len(r.points), func(i int) bool { return r.points[i].h >= h })
	if idx == len(r.points) {
		idx = 0
	}
	return r.points[idx].nodeID
}
