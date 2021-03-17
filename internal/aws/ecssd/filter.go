package ecssd

import (
	"fmt"
	"sort"

	"go.uber.org/zap"
)

type TaskFilter struct {
	looger   *zap.Logger
	matchers map[MatcherType][]Matcher
}

type TaskFilterOptions struct {
	Logger   *zap.Logger
	Matchers map[MatcherType][]Matcher
}

func NewTaskFilter(c Config, opts TaskFilterOptions) (*TaskFilter, error) {
	return &TaskFilter{
		looger:   opts.Logger,
		matchers: opts.Matchers,
	}, nil
}

// Filter run all the matchers and return all the tasks that including at least one matched container.
func (f *TaskFilter) Filter(tasks []*Task) ([]*Task, error) {
	matched := make(map[MatcherType][]*MatchResult)
	merr := newMulti()
	for tpe, matchers := range f.matchers {
		for index, matcher := range matchers {
			res, err := matchContainers(tasks, matcher, index)
			// NOTE: we continue the loop even if there is error because it could some tasks has invalid labels.
			// matchCotnainers always return non nil result even if there are errors during matching.
			if err != nil {
				merr.Append(fmt.Errorf("matcher failed with type %s index %d: %w", tpe, index, err))
			}

			f.looger.Debug("matched",
				zap.String("MatcherType", tpe.String()), zap.Int("MatcherIndex", index),
				zap.Int("Tasks", len(tasks)), zap.Int("MatchedTasks", len(res.Tasks)),
				zap.Int("MatchedContainers", len(res.Containers)))
			matched[tpe] = append(matched[tpe], res)
		}
	}

	matchedTasks := make(map[int]bool)
	for _, tpe := range matcherOrders() {
		for _, res := range matched[tpe] {
			for _, container := range res.Containers {
				matchedTasks[container.TaskIndex] = true
				task := tasks[container.TaskIndex]
				task.AddMatchedContainer(container)
			}
		}
	}

	// Sort by task index so the output is consistent.
	var taskIndexes []int
	for k := range matchedTasks {
		taskIndexes = append(taskIndexes, k)
	}
	sort.Ints(taskIndexes)
	var sortedTasks []*Task
	for _, i := range taskIndexes {
		task := tasks[i]
		// Sort containers within a task
		sort.Slice(task.Matched, func(i, j int) bool {
			return task.Matched[i].ContainerIndex < task.Matched[j].ContainerIndex
		})
		sortedTasks = append(sortedTasks, task)
	}
	return sortedTasks, merr.ErrorOrNil()
}
