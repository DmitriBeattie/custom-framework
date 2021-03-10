package app

import (
	"errors"
	"fmt"
	"sync"
)

type dependencyID string

func createDependency(parentInstanceName, childInstanceName string) (dependencyID, dependency) {
	return dependencyID(
			fmt.Sprintf(
				"%s;;;;%s",
				parentInstanceName,
				childInstanceName),
		),
		dependency{
			parentInstanceName: parentInstanceName,
			childInstanceName:  childInstanceName,
		}
}

type Instance interface {
	Process(data interface{}) error
	Name() string
	DependsOn() []string
}

type dependency struct {
	parentInstanceName string
	childInstanceName  string
}

type Aggregator struct {
	instances []Instance
	dep       map[dependencyID]dependency
	log       Logger
}

func NewAggregator(log Logger, instances ...Instance) (*Aggregator, error) {
	//return nil, errors.New("Need to be implemented")

	agr := &Aggregator{
		instances: instances,
		dep:       make(map[dependencyID]dependency),
		log:       log,
	}

	nodes := make(map[string]bool, len(instances))

	for _, inst := range instances {
		nodes[inst.Name()] = false
	}

	processedNodes := 0
	for {
		if processedNodes == len(nodes) {
			break
		}

		newProcessedNodes := processedNodes

		for _, inst := range instances {
			isDependFromNotProcessedNode := false
			for _, dependsOn := range inst.DependsOn() {
				isParentInstanceProcessed, isParentInstanceExists := nodes[dependsOn]
				if !isParentInstanceExists {
					return nil, fmt.Errorf("%s not found", dependsOn)
				}
				if isParentInstanceProcessed {
					id, dep := createDependency(dependsOn, inst.Name())
					agr.dep[id] = dep
				} else {
					isDependFromNotProcessedNode = true
				}
			}

			if !isDependFromNotProcessedNode {
				newProcessedNodes++
				nodes[inst.Name()] = true
			}
		}

		if newProcessedNodes == processedNodes {
			return nil, errors.New("Unable to build aggregator. Check if cycles or unreachable dependencies exists")
		}

		processedNodes = newProcessedNodes
	}

	return agr, nil
}

func (a *Aggregator) pipeline() (listeners, notifiers map[string][]chan error) {
	listeners = make(map[string][]chan error, len(a.dep))
	notifiers = make(map[string][]chan error, len(a.dep))

	for _, dep := range a.dep {
		ch := make(chan error, 1)
		listeners[dep.childInstanceName] = append(listeners[dep.childInstanceName], ch)
		notifiers[dep.parentInstanceName] = append(notifiers[dep.parentInstanceName], ch)
	}

	return
}

func (a *Aggregator) Process(data interface{}) {
	list, ntf := a.pipeline()

	instanceWg := sync.WaitGroup{}
	instanceWg.Add(len(a.instances))

	for i := range a.instances {
		go func(inst Instance) {
			defer instanceWg.Done()

			var parentError error

			if chz, ok := list[inst.Name()]; ok {
				wait := sync.WaitGroup{}
				wait.Add(len(chz))

				for j := range chz {
					go func(ind int) {
						defer wait.Done()

						err := <-chz[ind]
						if err != nil {
							parentError = err
						}
					}(j)
				}

				wait.Wait()
			}

			if parentError == nil {
				if err := inst.Process(data); err != nil {
					parentError = err
					a.log.Error(err)
				}
			}

			for _, notify := range ntf[inst.Name()] {
				notify <- parentError
				close(notify)
			}
		}(a.instances[i])
	}

	instanceWg.Wait()
}
