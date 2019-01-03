package consistenthash

import (
	"github.com/jtejido/hermes/t1ha"
	"math"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
)

// Consistent hashing with bounded loads
// https://research.googleblog.com/2017/04/consistent-hashing-with-bounded-loads.html
// https://arxiv.org/pdf/1608.01350.pdf
type Host struct {
	Name string
	Load uint64
}

func (h *Host) GetName() string {
	return h.Name
}

func (h *Host) GetLoad() uint64 {
	return h.Load
}

type Map struct {
	hosts     map[uint64]string
	keys      []uint64
	loadMap   map[string]*Host
	totalLoad uint64
	replicas  int
	sync.RWMutex
}

func New(replicas int) *Map {
	return &Map{
		replicas: replicas,
		hosts:    make(map[uint64]string),
		loadMap:  make(map[string]*Host),
	}
}

func (m *Map) IsEmpty() bool {
	return len(m.hosts) == 0
}

func (m *Map) Add(hosts ...string) {
	m.Lock()
	defer m.Unlock()

	for _, host := range hosts {
		m.loadMap[host] = &Host{Name: host, Load: 0}
		for i := 0; i < m.replicas; i++ {
			hash := hasher([]byte(strconv.Itoa(i)+host), 0)
			m.hosts[hash] = host
			m.keys = append(m.keys, hash)
		}
	}

	sort.Slice(m.keys, func(i int, j int) bool {
		if m.keys[i] < m.keys[j] {
			return true
		}
		return false
	})
}

func (m *Map) Get(key string) string {
	m.RLock()
	defer m.RUnlock()

	if m.IsEmpty() {
		return ""
	}

	hash := hasher([]byte(key), 0)
	idx := m.search(hash)

	i := idx
	for {
		host := m.hosts[m.keys[i]]
		if m.isLoadable(host) {
			return host
		}
		i++
		if i >= len(m.hosts) {
			i = 0
		}
	}
}

// incrementing a host's load. Use This when setting.
func (m *Map) Increment(host string) {

	if _, ok := m.loadMap[host]; !ok {
		return
	}

	atomic.AddUint64(&m.loadMap[host].Load, 1)
	atomic.AddUint64(&m.totalLoad, 1)
}

// decrementing a host's load. Use this when evicting.
func (m *Map) Decrement(host string) {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.loadMap[host]; !ok {
		return
	}

	if m.loadMap[host].Load > 0 {
		atomic.AddUint64(&m.loadMap[host].Load, ^uint64(0))
	}

	if m.totalLoad > 0 {
		atomic.AddUint64(&m.totalLoad, ^uint64(0))
	}

	return
}

func (m *Map) GetLoad(host string) (value uint64) {
	m.Lock()
	defer m.Unlock()

	v, ok := m.loadMap[host]
	if !ok {
		return
	}

	return v.Load
}

func (m *Map) search(key uint64) int {
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= key
	})

	if idx >= len(m.keys) {
		idx = 0
	}
	return idx
}

func (m *Map) GetHosts() (hosts []*Host) {
	m.RLock()
	defer m.RUnlock()
	for _, v := range m.loadMap {
		hosts = append(hosts, v)
	}
	return hosts
}

func (m *Map) GetLoads() map[string]uint64 {
	loads := map[string]uint64{}

	for k, v := range m.loadMap {
		loads[k] = v.Load
	}
	return loads
}

func (m *Map) isLoadable(host string) bool {
	if m.totalLoad < 0 {
		m.totalLoad = 0
	}

	var avgLoadPerNode float64
	avgLoadPerNode = float64((m.totalLoad + 1) / uint64(len(m.loadMap)))
	if avgLoadPerNode == 0 {
		avgLoadPerNode = 1
	}
	// Vimeo's Value
	avgLoadPerNode = math.Ceil(avgLoadPerNode * 1.25)

	b, ok := m.loadMap[host]
	if !ok {
		return false
	}

	if float64(b.Load)+1 <= avgLoadPerNode {
		return true
	}

	return false
}

func hasher(value []byte, seed uint64) uint64 {
	hasher := t1ha.New64WithSeed(seed)
	hasher.Write(value)
	return hasher.Sum64()
}
