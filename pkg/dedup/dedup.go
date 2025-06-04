package dedup

// Deduper tracks recently seen IDs to avoid processing duplicates.
type Deduper struct {
	capacity int
	set      map[string]struct{}
	order    []string
}

// New creates a Deduper with the given capacity.
func New(capacity int) *Deduper {
	if capacity <= 0 {
		capacity = 1
	}
	return &Deduper{
		capacity: capacity,
		set:      make(map[string]struct{}),
	}
}

// Seen reports whether id has been seen before. If not, it records the id and
// returns false. Once the capacity is exceeded, the oldest ID is evicted.
func (d *Deduper) Seen(id string) bool {
	if _, ok := d.set[id]; ok {
		return true
	}
	d.set[id] = struct{}{}
	d.order = append(d.order, id)
	if len(d.order) > d.capacity {
		oldest := d.order[0]
		d.order = d.order[1:]
		delete(d.set, oldest)
	}
	return false
}
