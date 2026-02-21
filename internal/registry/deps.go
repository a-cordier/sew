package registry

import (
	"errors"
	"fmt"
)

// Validate checks that all required components exist and that there are no cycles.
func Validate(components []Component) error {
	nameToComp := make(map[string]Component)
	for _, c := range components {
		nameToComp[c.Name] = c
	}
	for _, c := range components {
		for _, req := range c.Requires {
			if _, ok := nameToComp[req.Component]; !ok {
				return fmt.Errorf("component %q requires unknown component %q", c.Name, req.Component)
			}
		}
	}
	_, err := TopoSort(components)
	return err
}

// TopoSort returns components in install order (dependencies first).
func TopoSort(components []Component) ([]Component, error) {
	nameToComp := make(map[string]Component)
	for _, c := range components {
		nameToComp[c.Name] = c
	}

	inDegree := make(map[string]int)
	adj := make(map[string][]string)
	for name := range nameToComp {
		inDegree[name] = 0
	}
	for _, c := range components {
		for _, req := range c.Requires {
			adj[req.Component] = append(adj[req.Component], c.Name)
			inDegree[c.Name]++
		}
	}

	var queue []string
	for name, d := range inDegree {
		if d == 0 {
			queue = append(queue, name)
		}
	}

	var order []Component
	for len(queue) > 0 {
		name := queue[0]
		queue = queue[1:]
		order = append(order, nameToComp[name])
		for _, next := range adj[name] {
			inDegree[next]--
			if inDegree[next] == 0 {
				queue = append(queue, next)
			}
		}
	}

	if len(order) != len(components) {
		return nil, errors.New("dependency cycle detected among components")
	}
	return order, nil
}
