package api

type Scope []string

// CheckScope checks whether target scope has all required scopes
func (a *DefaultAPI) CheckScope(target Scope, required Scope) bool {
	var ok int

	for _, s := range required {
		for _, es := range target {
			if es == s {
				ok++
				break
			}
		}
	}

	return ok == len(required)
}
