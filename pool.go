package cobe

import (
	"fmt"
)

// Implements a connection pool for cobe brains.
type BrainPool struct {
	Brains map[string]*Cobe2Brain
}

func NewBrainPool() BrainPool {
	return BrainPool{Brains: make(map[string]*Cobe2Brain)}
}

// Add and initialize a connection to the pool.
func (cp *BrainPool) Add(name string) error {
	b, err := OpenCobe2Brain(name)
	if err != nil {
		return err
	}
	cp.Brains[name] = b
	return nil
}

// Retrieves any/first connection found in pool.
func (cp *BrainPool) First() *Cobe2Brain {
	for _, v := range cp.Brains {
		return v
	}
	return nil
}

// Retrieves a connection by its name.
func (cp *BrainPool) Get(name string) (*Cobe2Brain, error) {
	for k, v := range cp.Brains {
		if k == name {
			return v, nil
		}
	}
	return nil, fmt.Errorf("No such brain %s found in pool.", name)
}
