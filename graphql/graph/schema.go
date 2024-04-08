package graph

import (
	"fmt"

	"github.com/graphql-go/graphql"
)

func NewSchema() (*graphql.Schema, error) {
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: newQueryType(),
	})
	if err != nil {
		return nil, fmt.Errorf("new schema: %w", err)
	}
	return &schema, err
}

func newQueryType() *graphql.Object {
	var taskType *graphql.Object
	var projectType *graphql.Object

	taskType = newTaskType()
	projectType = newProjectType()

	taskType.AddFieldConfig("project", &graphql.Field{
		Name:    "Project",
		Type:    graphql.NewNonNull(projectType),
		Resolve: GetProjectForTask,
	})
	projectType.AddFieldConfig("tasks", &graphql.Field{
		Name:    "Tasks",
		Type:    graphql.NewList(graphql.NewNonNull(taskType)),
		Resolve: GetTasksOfProject,
	})

	return graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"projects": &graphql.Field{
				Type:    graphql.NewList(graphql.NewNonNull(projectType)),
				Resolve: GetProjects,
			},
			"tasks": &graphql.Field{
				Type: graphql.NewList(graphql.NewNonNull(taskType)),
				Args: graphql.FieldConfigArgument{
					"projectID": &graphql.ArgumentConfig{
						Type:         graphql.String,
						DefaultValue: "",
					},
				},
				Resolve: GetTasks,
			},
		},
	})
}

func newProjectType() *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: "Project",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
			},
			"name": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
			},
		},
	})
}

func newTaskType() *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: "Task",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
			},
			"content": &graphql.Field{
				Type: graphql.NewNonNull(graphql.String),
			},
		},
	})
}
