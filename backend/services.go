package main

import (
	"fmt"

	"github.com/graphql-go/graphql"
)

func ListDevelopers(repository DeveloperRepository) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		developers, err := repository.GetAll(p.Context)
		if err != nil {
			return nil, fmt.Errorf("getting all developers: %w", err)
		}

		return developers, nil
	}
}

func DeveloperDetails(repository DeveloperRepository) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		id, ok := p.Args["id"].(int)
		if !ok {
			return nil, nil
		}

		developer, err := repository.GetByID(p.Context, int64(id))
		if err != nil {
			return nil, fmt.Errorf("getting developer ID %d: %w", id, err)
		}

		return developer, nil
	}
}

func CreateDeveloper(repository DeveloperRepository) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		stack := toSlice[string](p.Args["stack"])
		newDeveloper := Developer{
			FirstName: p.Args["first_name"].(string),
			LastName:  p.Args["last_name"].(string),
			GithubURL: p.Args["github_url"].(string),
			Stack:     stack,
		}

		dev, err := repository.Create(p.Context, newDeveloper)
		if err != nil {
			return nil, fmt.Errorf("creating new developer: %w", err)
		}

		return dev, nil
	}
}

func UpdateDeveloper(repository DeveloperRepository) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		stack := toSlice[string](p.Args["stack"])
		developer := Developer{
			ID:        int64(p.Args["id"].(int)),
			FirstName: p.Args["first_name"].(string),
			LastName:  p.Args["last_name"].(string),
			GithubURL: p.Args["github_url"].(string),
			Stack:     stack,
		}

		dev, err := repository.Update(p.Context, developer)
		if err != nil {
			return nil, fmt.Errorf("updating developer: %w", err)
		}

		return dev, nil
	}
}

func toSlice[T any](slice interface{}) []T {
	sl, isSlice := slice.([]interface{})
	if !isSlice {
		return make([]T, 0)
	}
	converted := make([]T, 0, len(sl))
	for _, s := range sl {
		converted = append(converted, s.(T))
	}
	return converted
}
