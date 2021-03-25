package store

type InMemoryStore struct {
	Providers []Provider
}

func (s InMemoryStore) GetProvider(name string) (Provider, error) {
	for i := range s.Providers {
		if s.Providers[i].Name == name {
			return s.Providers[i], nil
		}
	}
	return Provider{Name: name}, ErrorNotFound
}

func (s *InMemoryStore) SetProvider(p Provider) (Provider, error) {
	for i := range s.Providers {
		if s.Providers[i].Name == p.Name {
			s.Providers[i] = p
			return p, nil
		}
	}
	s.Providers = append(s.Providers, p)
	return p, nil
}

func (s InMemoryStore) GetProviderList() ([]Provider, error) {
	l := make([]Provider, len(s.Providers))
	copy(l, s.Providers)
	return l, nil
}
