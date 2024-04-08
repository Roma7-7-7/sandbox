package graph

import (
	"context"

	"github.com/graphql-go/graphql"

	"github.com/Roma7-7-7/sandbox/graphql/pkg/todoist"
)

type TodoistClient interface {
	GetProjects(ctx context.Context) ([]todoist.Project, error)
	GetProject(ctx context.Context, projectID string) (*todoist.Project, error)
	GetTasks(ctx context.Context, projectID string) ([]todoist.Task, error)
	GetTask(ctx context.Context, taskID string) (*todoist.Task, error)
}

func WithResolveContext(ctx context.Context, client TodoistClient) context.Context {
	return context.WithValue(ctx, resolveContextKey{}, &resolveContext{
		client:  client,
		loaders: newLoaders(client),
	})
}

func GetProjects(p graphql.ResolveParams) (interface{}, error) {
	var (
		resolveCtx = p.Context.Value(resolveContextKey{}).(*resolveContext)
		key        = newResolveKey("", resolveCtx.client)
		thunk      = resolveCtx.loaders.GetProjects.Load(p.Context, key)
	)

	type result struct {
		projects []*Project
		err      error
	}
	ch := make(chan result, 1)

	go func() {
		defer close(ch)
		r, err := thunk()
		if err != nil {
			ch <- result{err: err}
			return
		}
		projects := r.([]*Project)
		ch <- result{projects: projects}
	}()

	return func() (interface{}, error) {
		r := <-ch
		if r.err != nil {
			return nil, r.err
		}
		return r.projects, nil
	}, nil
}

func GetProjectForTask(p graphql.ResolveParams) (interface{}, error) {
	task := p.Source.(*Task)
	resolveCtx := p.Context.Value(resolveContextKey{}).(*resolveContext)
	key := newResolveKey(task.ProjectID, resolveCtx.client)
	thunk := resolveCtx.loaders.GetProjects.Load(p.Context, key)

	type result struct {
		project *Project
		err     error
	}
	ch := make(chan result, 1)

	go func() {
		defer close(ch)
		r, err := thunk()
		if err != nil {
			ch <- result{err: err}
			return
		}
		projects := r.([]*Project)
		ch <- result{project: projects[0]}
	}()

	return func() (interface{}, error) {
		r := <-ch
		if r.err != nil {
			return nil, r.err
		}
		return r.project, nil
	}, nil
}

func GetTasks(p graphql.ResolveParams) (interface{}, error) {
	projectID, _ := p.Args["projectID"].(string)
	return getTasksByProjectID(p, projectID)
}

func GetTasksOfProject(p graphql.ResolveParams) (interface{}, error) {
	project := p.Source.(*Project)
	return getTasksByProjectID(p, project.ID)
}

func getTasksByProjectID(p graphql.ResolveParams, projectID string) (interface{}, error) {
	resolveCtx := p.Context.Value(resolveContextKey{}).(*resolveContext)
	key := newResolveKey(projectID, resolveCtx.client)
	thunk := resolveCtx.loaders.GetTasks.Load(p.Context, key)

	type result struct {
		tasks []*Task
		err   error
	}
	ch := make(chan result, 1)

	go func() {
		defer close(ch)
		r, err := thunk()
		if err != nil {
			ch <- result{err: err}
			return
		}
		tasks := r.([]*Task)
		ch <- result{tasks: tasks}
	}()

	return func() (interface{}, error) {
		r := <-ch
		if r.err != nil {
			return nil, r.err
		}
		return r.tasks, nil
	}, nil
}
