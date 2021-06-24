// Copyright 2021 The sacloud Authors
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
	"fmt"

	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/shivamMg/ppds/tree"
)

type Graph struct {
	defGroups *ResourceDefGroups
	children  []tree.Node
}

func NewGraph(defGroups *ResourceDefGroups) *Graph {
	return &Graph{
		defGroups: defGroups,
	}
}

func (n *Graph) Data() interface{} {
	return "Sacloud AutoScaler"
}

func (n *Graph) Children() []tree.Node {
	return n.children
}

func (g *Graph) Tree(ctx *RequestContext, apiClient sacloud.APICaller) (string, error) {
	g.children = []tree.Node{}
	for _, group := range g.defGroups.All() {
		groupNode := &GroupNode{name: group.name}
		for _, def := range group.ResourceDefs {
			nodes, err := g.nodes(ctx, apiClient, def)
			if err != nil {
				return "", err
			}
			groupNode.children = append(groupNode.children, nodes...)
		}
		g.children = append(g.children, groupNode)
	}
	return tree.SprintHrn(g), nil
}

func (g *Graph) nodes(ctx *RequestContext, apiClient sacloud.APICaller, def ResourceDefinition) ([]tree.Node, error) {
	resources, err := def.Compute(ctx, apiClient)
	if err != nil {
		return nil, err
	}
	var nodes []tree.Node
	for _, r := range resources {
		node := &GraphNode{resource: r}
		for _, child := range def.Children() {
			children, err := g.nodes(ctx, apiClient, child)
			if err != nil {
				return nil, err
			}
			node.children = append(node.children, children...)
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

type GroupNode struct {
	name     string
	children []tree.Node
}

func (n *GroupNode) Data() interface{} {
	return fmt.Sprintf("Group: %s", n.name)
}

func (n *GroupNode) Children() []tree.Node {
	return n.children
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
