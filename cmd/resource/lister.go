package resource

import (
	"open-cluster-management.io/sdk-go/pkg/cloudevents/generic"
	"open-cluster-management.io/sdk-go/pkg/cloudevents/generic/types"
)

type ResourceLister struct{}

var _ generic.Lister[*Resource] = &ResourceLister{}

func (resLister *ResourceLister) List(listOpts types.ListOptions) ([]*Resource, error) {
	return store.List(listOpts.ClusterName), nil
}
