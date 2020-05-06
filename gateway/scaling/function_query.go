// Copyright (c) OpenFaaS Author(s). All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package scaling

import "fmt"

type CachedFunctionQuery struct {
	cache            FunctionCacher
	serviceQuery     ServiceQuery
	emptyAnnotations map[string]string
}

func NewCachedFunctionQuery(cache FunctionCacher, serviceQuery ServiceQuery) FunctionQuery {
	return &CachedFunctionQuery{
		cache:            cache,
		serviceQuery:     serviceQuery,
		emptyAnnotations: map[string]string{},
	}
}

func (c *CachedFunctionQuery) GetAnnotations(name string, namespace string) (annotations map[string]string, err error) {
	res, err := c.Get(name, namespace)
	if err != nil {
		return c.emptyAnnotations, err
	}

	if res.Annotations == nil {
		return c.emptyAnnotations, nil
	}
	return *res.Annotations, nil
}

func (c *CachedFunctionQuery) Get(fn string, ns string) (ServiceQueryResponse, error) {

	query, hit := c.cache.Get(fn, ns)
	if !hit {

		// If there is a cache miss, then fetch the value from the provider API
		queryResponse, err := c.serviceQuery.GetReplicas(fn, ns)
		if err != nil {
			return ServiceQueryResponse{}, err
		}
		c.cache.Set(fn, ns, queryResponse)
	} else {
		return query, nil
	}

	// At this point the value almost certainly must be present, so if not
	// return an error.
	query, hit = c.cache.Get(fn, ns)
	if !hit {
		return ServiceQueryResponse{}, fmt.Errorf("error with cache key: %s", fn+"."+ns)
	}

	return query, nil
}

type FunctionQuery interface {
	Get(name string, namespace string) (ServiceQueryResponse, error)
	GetAnnotations(name string, namespace string) (annotations map[string]string, err error)
}
