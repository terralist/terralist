package slug

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrInvalidSlug = errors.New("Invalid slug format")
)

type Slug struct {
	Namespace string
	Name      string
	Provider  string
}

func (s *Slug) String() string {
	if s.Provider == "" {
		return fmt.Sprintf("%s/%s", s.Namespace, s.Name)
	}

	return fmt.Sprintf("%s/%s/%s", s.Namespace, s.Name, s.Provider)
}

func Parse(slug string) (*Slug, error) {
	tokens := strings.Split(slug, "/")

	// hashicorp/null
	// hashicorp/vpc/aws

	// Providers have 2 tokens in their slug
	if len(tokens) == 2 {
		return &Slug{
			Namespace: tokens[0],
			Name:      tokens[1],
		}, nil
	}

	// Modules have 3 tokens in their slug
	if len(tokens) == 3 {
		return &Slug{
			Namespace: tokens[0],
			Name:      tokens[1],
			Provider:  tokens[3],
		}, nil
	}

	// Anything else is not a valid slug
	return nil, ErrInvalidSlug
}

func MustParse(slug string) *Slug {
	s, err := Parse(slug)
	if err != nil {
		panic(err)
	}

	return s
}
