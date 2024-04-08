package graph

import (
	"context"
	"fmt"
	"sync"

	"github.com/graph-gophers/dataloader"
)

type (
	loader interface {
		Load(context.Context, dataloader.Key) dataloader.Thunk
	}

	loaders struct {
		GetProjects loader
		GetTasks    loader
	}

	resolveContextKey struct{}

	resolveContext struct {
		client  TodoistClient
		loaders loaders
	}

	resolveKey struct {
		key    string
		client TodoistClient
	}
)

func newLoaders(client TodoistClient) loaders {
	return loaders{
		GetProjects: dataloader.NewBatchedLoader(LoadProjects(client)),
		GetTasks:    dataloader.NewBatchedLoader(LoadTasks(client)),
	}
}

func newResolveKey(key string, client TodoistClient) *resolveKey {
	return &resolveKey{key: key, client: client}
}

func (k *resolveKey) String() string {
	return k.key
}

func (k *resolveKey) Raw() interface{} {
	return k
}

func LoadProjects(client TodoistClient) func(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
	return func(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
		var (
			wg      = sync.WaitGroup{}
			results = make(map[int][]*dataloader.Result, len(keys))
			mx      = sync.Mutex{}
		)

		for i, key := range keys {
			wg.Add(1)
			go func(i int, key dataloader.Key) {
				defer wg.Done()
				projects, err := client.GetProjects(ctx)
				if err != nil {
					results[i] = []*dataloader.Result{{Error: fmt.Errorf("get projects: %w", err)}}
					return
				}

				converted := make([]*Project, len(projects))
				for i, p := range projects {
					project := &Project{
						ID:   p.ID,
						Name: p.Name,
					}
					converted[i] = project
				}
				mx.Lock()
				defer mx.Unlock()
				results[i] = []*dataloader.Result{{Data: converted}}
			}(i, key)
		}

		wg.Wait()

		res := make([]*dataloader.Result, len(keys))
		for i, r := range results {
			res[i] = r[0]
		}
		return res
	}
}

func LoadTasks(client TodoistClient) func(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
	return func(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
		var (
			wg      = sync.WaitGroup{}
			results = make(map[int][]*dataloader.Result, len(keys))
			mx      = sync.Mutex{}
		)

		for i, key := range keys {
			wg.Add(1)
			go func(i int, key dataloader.Key) {
				defer wg.Done()
				tasks, err := client.GetTasks(ctx, key.(*resolveKey).key)
				if err != nil {
					results[i] = []*dataloader.Result{{Error: fmt.Errorf("get tasks: %w", err)}}
					return
				}

				converted := make([]*Task, len(tasks))
				for i, t := range tasks {
					task := &Task{
						ID:        t.ID,
						Content:   t.Content,
						ProjectID: t.ProjectID,
					}
					converted[i] = task
				}
				mx.Lock()
				defer mx.Unlock()
				results[i] = []*dataloader.Result{{Data: converted}}
			}(i, key)
		}

		wg.Wait()

		res := make([]*dataloader.Result, len(keys))
		for i, r := range results {
			res[i] = r[0]
		}
		return res
	}
}
