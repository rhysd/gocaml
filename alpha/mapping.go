package alpha

type mapping struct {
	parent *mapping
	vars   map[string]string
}

func newMapping(parent *mapping) *mapping {
	return &mapping{
		parent,
		map[string]string{},
	}
}

func (m *mapping) add(from, to string) {
	m.vars[from] = to
}

func (m *mapping) resolve(name string) (string, bool) {
	if mapped, ok := m.vars[name]; ok {
		return mapped, true
	}
	if m.parent == nil {
		return "", false
	}
	return m.parent.resolve(name)
}
