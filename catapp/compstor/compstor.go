package compstor

import (
	"errors"
	"slices"
	"sync"

	"github.com/surkovvs/gocat/catapp/component"
)

var (
	ErrGroupAlreadyRegistered     = errors.New("group already registered")
	ErrGroupNotFound              = errors.New("group not found")
	ErrComponentAlreadyRegistered = errors.New("component already registered")
)

type groupNum int

type CompsStorage struct {
	mu           *sync.Mutex
	comps        map[component.Comp]groupNum
	groups       map[string]SequentialGroup
	groupCounter groupNum
}

type SequentialGroup struct {
	name  string
	num   groupNum
	comps []component.Comp
}

func NewCompsStorage() CompsStorage {
	return CompsStorage{
		mu:     &sync.Mutex{},
		comps:  make(map[component.Comp]groupNum),
		groups: make(map[string]SequentialGroup),
	}
}

func (cs *CompsStorage) AddComponent(groupName, compName string, comp component.Comp) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	_, ok := cs.comps[comp]
	if ok {
		return ErrComponentAlreadyRegistered
	}

	group, ok := cs.groups[groupName]
	if !ok {
		group = SequentialGroup{
			name: groupName,
			num:  cs.groupCounter,
		}
		cs.groupCounter++
	}
	group.comps = append(group.comps, comp)
	cs.groups[groupName] = group

	cs.comps[comp] = group.num
	return nil
}

func (cs *CompsStorage) AddGroup(groupName string) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	_, ok := cs.groups[groupName]
	if ok {
		return ErrGroupAlreadyRegistered
	}

	cs.groups[groupName] = SequentialGroup{
		name: groupName,
		num:  cs.groupCounter,
	}
	cs.groupCounter++
	return nil
}

func (cs *CompsStorage) GetOrderedGroupList() []SequentialGroup {
	cs.mu.Lock()
	groupList := make([]SequentialGroup, 0, len(cs.groups))
	for _, group := range cs.groups {
		groupList = append(groupList, group)
	}
	cs.mu.Unlock()
	slices.SortFunc(groupList, func(a, b SequentialGroup) int {
		return int(a.num - b.num)
	})
	return groupList
}

func (cs *CompsStorage) GetGroupByName(name string) (SequentialGroup, error) {
	cs.mu.Lock()
	group, ok := cs.groups[name]
	cs.mu.Unlock()
	if !ok {
		return group, ErrGroupNotFound
	}
	return group, nil
}

func (cs *CompsStorage) GetUnsortedShutdowners() []component.Comp {
	cs.mu.Lock()
	compList := make([]component.Comp, 0, len(cs.comps))
	for comp := range cs.comps {
		if comp.IsShutdowner() {
			compList = append(compList, comp)
		}
	}
	cs.mu.Unlock()
	return compList
}

func (sg SequentialGroup) GetName() string {
	return sg.name
}

func (sg SequentialGroup) GetComponents() []component.Comp {
	return sg.comps
}

// // TODO:
// func (cs *CompsStorage) GetComponentByName() {
// 	panic("not implemented")
// }

// // TODO:
// func (cs *CompsStorage) RemoveComponent() {
// 	panic("not implemented")
// }

// // TODO:
// func (cs CompsStorage) RemoveGroup() {
// 	panic("not implemented")
// }
