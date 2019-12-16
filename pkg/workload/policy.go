package workload

import (
	"fmt"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

type PolicySpec struct {
	Operation     string
	ResourceNames []string
	Parameters    map[string]string
}

// Policy configures a Task.
type Policy struct {
	// Resources is the resource types to operate on.
	Resources []schema.GroupVersionResource

	// Interval is the time to wait between operations.
	Interval time.Duration

	// Limit is the max number of operations per interval.
	Limit int

	// TODO(jpeach): Add Probability (of each operation occuring).
}

// ParsePolicySpec parses a policy specification string. A policy
// specification string is of the form:
//
//	OPERATION[:RESOURCE|PARAM=VALUE[,RESOURCE|PARAM=VALUE]...]
func ParsePolicySpec(specString string) (PolicySpec, error) {
	if specString == "" {
		return PolicySpec{}, fmt.Errorf("empty policy specification")
	}

	spec := PolicySpec{
		Parameters: make(map[string]string),
	}

	// Outer syntax is "$OP:$ARGS".
	parts := strings.SplitN(specString, ":", 2)
	spec.Operation = parts[0]

	switch len(parts) {
	case 1:
		// Just an operation, with no policy args.
		return spec, nil
	case 2:
		for _, arg := range strings.Split(parts[1], ",") {
			parts := strings.SplitN(arg, "=", 2)
			switch len(parts) {
			case 1:
				spec.ResourceNames = append(spec.ResourceNames, parts[0])
			case 2:
				spec.Parameters[parts[0]] = parts[1]
			}
		}

		return spec, nil
	default:
		panic("invalid split")
	}
}
