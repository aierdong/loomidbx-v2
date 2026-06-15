package plan

import (
	"sort"
	"strconv"
	"strings"
)

// PlannedTable stores one table in the final dependency execution order.
type PlannedTable struct {
	// ProjectTableID stores the ProjectTable identity from the graph node.
	ProjectTableID int64

	// TableID stores the Schema table identity from the graph node.
	TableID int64

	// ExecutionOrder stores the one-based stable order assigned by the topological sorter.
	ExecutionOrder int
}

// SortDependencyGraph returns a deterministic topological order for all graph nodes.
func SortDependencyGraph(graph DependencyGraph) ([]PlannedTable, []PlanIssue) {
	nodesByID := make(map[int64]GraphNode, len(graph.Nodes))
	indegree := make(map[int64]int, len(graph.Nodes))
	outgoing := make(map[int64][]int64, len(graph.Nodes))

	for _, node := range graph.Nodes {
		nodesByID[node.ProjectTableID] = node
		indegree[node.ProjectTableID] = 0
	}

	for _, edge := range graph.Edges {
		if _, exists := nodesByID[edge.FromProjectTableID]; !exists {
			continue
		}
		if _, exists := nodesByID[edge.ToProjectTableID]; !exists {
			continue
		}
		outgoing[edge.FromProjectTableID] = append(outgoing[edge.FromProjectTableID], edge.ToProjectTableID)
		indegree[edge.ToProjectTableID]++
	}

	for nodeID := range outgoing {
		sort.SliceStable(outgoing[nodeID], func(i, j int) bool {
			return compareGraphNodes(nodesByID[outgoing[nodeID][i]], nodesByID[outgoing[nodeID][j]]) < 0
		})
	}

	ready := make([]GraphNode, 0, len(graph.Nodes))
	for _, node := range graph.Nodes {
		if indegree[node.ProjectTableID] == 0 {
			ready = append(ready, node)
		}
	}
	sortGraphNodes(ready)

	ordered := make([]PlannedTable, 0, len(graph.Nodes))
	for len(ready) > 0 {
		node := ready[0]
		ready = ready[1:]
		ordered = append(ordered, PlannedTable{
			ProjectTableID: node.ProjectTableID,
			TableID:        node.TableID,
			ExecutionOrder: len(ordered) + 1,
		})

		for _, downstreamID := range outgoing[node.ProjectTableID] {
			indegree[downstreamID]--
			if indegree[downstreamID] == 0 {
				ready = append(ready, nodesByID[downstreamID])
			}
		}
		sortGraphNodes(ready)
	}

	if len(ordered) != len(graph.Nodes) {
		issues := detectCycleIssues(nodesByID, indegree, outgoing)
		issues = append(issues, NewPlanIssue(
			PlanErrorCodeUnsortableGraph,
			PlanStageTopologicalSort,
			"graph.nodes",
			"dependency graph cannot be sorted into a complete table order",
			true,
		))
		return nil, issues
	}

	return ordered, nil
}

func detectCycleIssues(nodesByID map[int64]GraphNode, indegree map[int64]int, outgoing map[int64][]int64) []PlanIssue {
	cycleNodeIDs := make([]int64, 0, len(nodesByID))
	for nodeID := range nodesByID {
		if indegree[nodeID] > 0 && canReachNode(nodeID, nodeID, outgoing, nil) {
			cycleNodeIDs = append(cycleNodeIDs, nodeID)
		}
	}
	sort.SliceStable(cycleNodeIDs, func(i, j int) bool {
		return compareGraphNodes(nodesByID[cycleNodeIDs[i]], nodesByID[cycleNodeIDs[j]]) < 0
	})

	visited := make(map[int64]bool, len(cycleNodeIDs))
	issues := make([]PlanIssue, 0)
	for _, nodeID := range cycleNodeIDs {
		if visited[nodeID] {
			continue
		}
		component := collectCycleComponent(nodeID, cycleNodeIDs, outgoing, visited)
		if len(component) == 0 {
			continue
		}
		issues = append(issues, NewPlanIssue(
			PlanErrorCodeCycleDetected,
			PlanStageTopologicalSort,
			"graph.cycles["+formatNodeIDList(component)+"]",
			"dependency graph contains a cycle among project table nodes",
			true,
		))
	}
	return issues
}

func collectCycleComponent(start int64, cycleNodeIDs []int64, outgoing map[int64][]int64, visited map[int64]bool) []int64 {
	cycleNodeSet := make(map[int64]bool, len(cycleNodeIDs))
	for _, nodeID := range cycleNodeIDs {
		cycleNodeSet[nodeID] = true
	}

	component := make([]int64, 0)
	queue := []int64{start}
	visited[start] = true
	for len(queue) > 0 {
		nodeID := queue[0]
		queue = queue[1:]
		component = append(component, nodeID)
		for _, downstreamID := range outgoing[nodeID] {
			if !cycleNodeSet[downstreamID] || visited[downstreamID] || !canReachNode(downstreamID, nodeID, outgoing, nil) {
				continue
			}
			visited[downstreamID] = true
			queue = append(queue, downstreamID)
		}
	}
	sort.SliceStable(component, func(i, j int) bool {
		return component[i] < component[j]
	})
	return component
}

func canReachNode(start int64, target int64, outgoing map[int64][]int64, seen map[int64]bool) bool {
	if seen == nil {
		seen = make(map[int64]bool)
	}
	if seen[start] {
		return false
	}
	seen[start] = true
	for _, downstreamID := range outgoing[start] {
		if downstreamID == target || canReachNode(downstreamID, target, outgoing, seen) {
			return true
		}
	}
	return false
}

func formatNodeIDList(nodeIDs []int64) string {
	values := make([]string, 0, len(nodeIDs))
	for _, nodeID := range nodeIDs {
		values = append(values, strconv.FormatInt(nodeID, 10))
	}
	return strings.Join(values, ",")
}

func sortGraphNodes(nodes []GraphNode) {
	sort.SliceStable(nodes, func(i, j int) bool {
		return compareGraphNodes(nodes[i], nodes[j]) < 0
	})
}

func compareGraphNodes(left GraphNode, right GraphNode) int {
	leftHasOrder := left.StableOrder != 0
	rightHasOrder := right.StableOrder != 0
	if leftHasOrder != rightHasOrder {
		if leftHasOrder {
			return -1
		}
		return 1
	}
	if leftHasOrder && left.StableOrder != right.StableOrder {
		if left.StableOrder < right.StableOrder {
			return -1
		}
		return 1
	}
	if left.ProjectTableID != right.ProjectTableID {
		if left.ProjectTableID < right.ProjectTableID {
			return -1
		}
		return 1
	}
	if left.TableID != right.TableID {
		if left.TableID < right.TableID {
			return -1
		}
		return 1
	}
	return 0
}
