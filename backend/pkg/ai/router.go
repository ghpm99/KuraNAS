package ai

import "fmt"

// Route binds a task type to a primary provider and an optional fallback.
type Route struct {
	Primary  Provider
	Fallback Provider
}

// Router implements the Strategy pattern, mapping each TaskType
// to a specific provider (with optional fallback).
type Router struct {
	routes map[TaskType]Route
}

// NewRouter creates an empty Router.
func NewRouter() *Router {
	return &Router{
		routes: make(map[TaskType]Route),
	}
}

// Register maps a task type to a primary provider.
func (r *Router) Register(taskType TaskType, primary Provider) {
	r.routes[taskType] = Route{Primary: primary}
}

// RegisterWithFallback maps a task type to a primary provider and a fallback.
func (r *Router) RegisterWithFallback(taskType TaskType, primary, fallback Provider) {
	r.routes[taskType] = Route{Primary: primary, Fallback: fallback}
}

// Resolve returns the route for the given task type.
func (r *Router) Resolve(taskType TaskType) (Route, error) {
	route, ok := r.routes[taskType]
	if !ok {
		return Route{}, fmt.Errorf("%w: %s", ErrNoProviderForTask, taskType)
	}
	return route, nil
}

// RegisteredTaskTypes returns all task types that have a provider configured.
func (r *Router) RegisteredTaskTypes() []TaskType {
	types := make([]TaskType, 0, len(r.routes))
	for t := range r.routes {
		types = append(types, t)
	}
	return types
}
