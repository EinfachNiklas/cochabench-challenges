'use strict';

/**
 * Finds the shortest path in a weighted graph with budget constraints.
 *
 * The graph has edges with two metrics:
 * - 'distance': The distance (metric to minimize)
 * - 'cost': The cost (budget constraint)
 *
 * @param {Object} graph - Graph as adjacency list
 *   Format: { nodeId: [{ to: nodeId, distance: number, cost: number }, ...], ... }
 * @param {string|number} start - Start node
 * @param {string|number} end - Target node
 * @param {number} maxCost - Maximum cost budget
 * @returns {Object|null} { path: [nodes], totalDistance: number, totalCost: number }
 *                        or null if no path exists
 *
 * Constraints:
 * - Total cost must not exceed maxCost
 * - Minimize total distance while staying within the budget
 * - Cycles must be avoided
 * - Return null if no path exists or budget is insufficient
 */
function findConstrainedPath(graph, start, end, maxCost) {
	// TODO: Implement the algorithm
	return null;
}

/**
 * Computes all possible paths between two nodes and returns the
 * k shortest paths that stay within the budget.
 *
 * @param {Object} graph - Graph as adjacency list (see above)
 * @param {string|number} start - Start node
 * @param {string|number} end - Target node
 * @param {number} maxCost - Maximum cost budget
 * @param {number} k - Number of shortest paths to return
 * @returns {Array<Object>} Sorted list of the k shortest paths
 *   Format: [{ path: [...], totalDistance: number, totalCost: number }, ...]
 */
function findKShortestPaths(graph, start, end, maxCost, k) {
	// TODO: Implement the algorithm (challenging!)
	return [];
}

/**
 * Validates whether the graph is valid:
 * - All 'to' nodes exist as keys in the graph
 * - All distance and cost values are non-negative numbers
 * - No self-loops (node pointing to itself)
 *
 * @param {Object} graph - The graph to validate
 * @returns {boolean} true if valid, false otherwise
 */
function isValidGraph(graph) {
	// TODO: Implement the validation
	return false;
}

/**
 * Finds the shortest path that passes through specific waypoints.
 * Waypoints must be visited in the given order.
 *
 * @param {Object} graph - Graph as adjacency list
 * @param {string|number} start - Start node
 * @param {Array<string|number>} waypoints - Nodes to visit in order
 * @param {string|number} end - Target node
 * @param {number} maxCost - Maximum cost budget
 * @returns {Object|null} { path: [...], totalDistance: number, totalCost: number }
 */
function findPathWithWaypoints(graph, start, waypoints, end, maxCost) {
	// TODO: Implement the algorithm
	return null;
}

module.exports = {
	findConstrainedPath,
	findKShortestPaths,
	isValidGraph,
	findPathWithWaypoints
};
