// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"fmt"
	"io"
	"strconv"

	"github.com/smartxworks/lynx/plugin/tower/pkg/schema"
)

type LabelEvent struct {
	Mutation       MutationType            `json:"mutation"`
	Node           *schema.Label           `json:"node"`
	PreviousValues *schema.ObjectReference `json:"previousValues"`
}

type Login struct {
	Token string `json:"token"`
}

type LoginInput struct {
	Password string     `json:"password"`
	Source   UserSource `json:"source"`
	Username string     `json:"username"`
}

type VMEvent struct {
	Mutation       MutationType            `json:"mutation"`
	Node           *schema.VM              `json:"node"`
	PreviousValues *schema.ObjectReference `json:"previousValues"`
}

type MutationType string

const (
	MutationTypeCreated MutationType = "CREATED"
	MutationTypeDeleted MutationType = "DELETED"
	MutationTypeUpdated MutationType = "UPDATED"
)

var AllMutationType = []MutationType{
	MutationTypeCreated,
	MutationTypeDeleted,
	MutationTypeUpdated,
}

func (e MutationType) IsValid() bool {
	switch e {
	case MutationTypeCreated, MutationTypeDeleted, MutationTypeUpdated:
		return true
	}
	return false
}

func (e MutationType) String() string {
	return string(e)
}

func (e *MutationType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = MutationType(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid MutationType", str)
	}
	return nil
}

func (e MutationType) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type UserSource string

const (
	UserSourceLdap  UserSource = "LDAP"
	UserSourceLocal UserSource = "LOCAL"
)

var AllUserSource = []UserSource{
	UserSourceLdap,
	UserSourceLocal,
}

func (e UserSource) IsValid() bool {
	switch e {
	case UserSourceLdap, UserSourceLocal:
		return true
	}
	return false
}

func (e UserSource) String() string {
	return string(e)
}

func (e *UserSource) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = UserSource(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid UserSource", str)
	}
	return nil
}

func (e UserSource) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
