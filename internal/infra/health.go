package infra

import "context"

type DependencyCheck struct {
	Name  string
	Check func(ctx context.Context) error
}

type DependencyStatus struct {
	Name  string
	Ready bool
	Error string
}

type ReadinessResult struct {
	Status       string
	Ready        bool
	Dependencies []DependencyStatus
}

func CheckDependencies(ctx context.Context, checks []DependencyCheck) ReadinessResult {
	result := ReadinessResult{
		Status:       "ready",
		Ready:        true,
		Dependencies: make([]DependencyStatus, 0, len(checks)),
	}

	for _, check := range checks {
		status := DependencyStatus{
			Name:  check.Name,
			Ready: true,
		}

		if err := check.Check(ctx); err != nil {
			status.Ready = false
			status.Error = err.Error()
			result.Ready = false
			result.Status = "not_ready"
		}

		result.Dependencies = append(result.Dependencies, status)
	}

	return result
}
