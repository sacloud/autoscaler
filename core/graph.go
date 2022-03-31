// Copyright 2021-2022 The sacloud/autoscaler Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package core

import (
	"github.com/sacloud/iaas-api-go"
	"github.com/shivamMg/ppds/tree"
)

type Graph struct {
	resources ResourceDefinitions
	children  []tree.Node
}

func NewGraph(resources ResourceDefinitions) *Graph {
	return &Graph{
		resources: resources,
	}
}

func (g *Graph) Data() interface{} {
	return "Sacloud AutoScaler"
}

func (g *Graph) Children() []tree.Node {
	return g.children
}

func (g *Graph) Tree(ctx *RequestContext, apiClient iaas.APICaller) (string, error) {
	g.children = []tree.Node{}
	for _, def := range g.resources {
		nodes, err := g.nodes(ctx, apiClient, def)
		if err != nil {
			return "", err
		}
		g.children = append(g.children, nodes...)
	}
	return tree.SprintHrn(g), nil
}

func (g *Graph) nodes(ctx *RequestContext, apiClient iaas.APICaller, def ResourceDefinition) ([]tree.Node, error) {
	resources, err := def.Compute(ctx, apiClient)
	if err != nil {
		return nil, err
	}

	var parentNode *GraphNode
	var nodes []tree.Node
	for i, r := range resources {
		if i == 0 {
			if v, ok := r.(ChildResource); ok {
				parent := v.Parent()
				if parent != nil {
					parentNode = &GraphNode{resource: parent}
				}
			}
		}
		nodes = append(nodes, &GraphNode{resource: r})
	}

	if parentNode != nil {
		parentNode.children = nodes
		return []tree.Node{parentNode}, nil
	}
	return nodes, nil
}

type GraphNode struct {
	resource Resource
	children []tree.Node
}

func (n *GraphNode) Data() interface{} {
	return n.resource.String()
}

func (n *GraphNode) Children() []tree.Node {
	return n.children
}
