package controller

import (
	"github.com/cloudfoundry/security-operator/pkg/controller/podsecuritypolicybinding"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, podsecuritypolicybinding.Add)
}
