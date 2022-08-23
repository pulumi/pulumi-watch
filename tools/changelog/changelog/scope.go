package changelog

import (
	"fmt"
	"strings"
)

// Structured scope annotation, with YAML marshaling to and from a string.
type Scope struct {
	Primary   string
	SubScopes []string
}

func (s Scope) String() string {
	if len(s.SubScopes) == 0 {
		return s.Primary
	}
	return s.Primary + "/" + strings.Join(s.SubScopes, ",")
}

func (s Scope) MarshalYAML() (any, error) {
	return s.String(), nil
}

func (s *Scope) UnmarshalYAML(unmarshal func(any) error) error {
	var value string
	if err := unmarshal(&value); err != nil {
		return err
	}

	primary, subscopes, err := ParseScope(Config{}, value, true)
	if err != nil {
		return err
	}
	s.Primary = primary
	s.SubScopes = subscopes

	return nil
}

// ParseScope ensures a scope string is in the permitted list and if so, returns the scope and permitted subscopes
// permitted. No validation on allowed scopes/subscopes is done if "force" is set.
func ParseScope(config Config, value string, force bool) (string, []string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return "", nil, nil
	}

	scope, rest, _ := strings.Cut(value, "/")
	permittedSubs, isPermittedScope := config.Scopes.Get(scope)

	if len(rest) > 0 {
		var subs []string
		strings.Split(rest, "")

		subs = strings.Split(rest, ",")

		for idx, v := range subs {
			subs[idx] = strings.ToLower(strings.TrimSpace(v))
		}

		for _, sub := range subs {
			allowed := false
			for _, permittedSub := range permittedSubs {
				if sub == permittedSub {
					allowed = true
					break
				}
			}
			if !allowed && !force {
				if len(config.Scopes.Items) == 0 || !isPermittedScope {
					return scope, subs, fmt.Errorf("invalid subscope %q found, scope %v expects none", sub, scope)
				}

				return scope, subs, fmt.Errorf("invalid subscope %q found, "+
					"expected one of: %v; or use the force option to override", sub, strings.Join(permittedSubs, ", "))
			}
		}

		if !isPermittedScope && !force {
			return scope, subs, fmt.Errorf("invalid scope found, use --help to list available scopes")
		}

		return scope, subs, nil
	}

	return scope, nil, nil
}
