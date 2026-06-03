package ai

import "fmt"

// Route binds a task type to a primary provider and an ordered list of fallbacks.
type Route struct {
	Primary   Provider
	Fallbacks []Provider
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

// RegisterWithFallback maps a task type to a primary provider and a single fallback.
func (r *Router) RegisterWithFallback(taskType TaskType, primary, fallback Provider) {
	r.routes[taskType] = Route{Primary: primary, Fallbacks: []Provider{fallback}}
}

// RegisterChain maps a task type to an ordered chain of providers: the first
// is the primary and the remaining ones are tried, in order, as fallbacks.
func (r *Router) RegisterChain(taskType TaskType, providers ...Provider) {
	if len(providers) == 0 {
		return
	}
	r.routes[taskType] = Route{Primary: providers[0], Fallbacks: providers[1:]}
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
