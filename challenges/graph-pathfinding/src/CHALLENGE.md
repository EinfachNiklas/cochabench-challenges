# Graph Pathfinding with Constraints

## Task

Implement several advanced pathfinding algorithms for weighted graphs with budget constraints.

All functions operate on a graph in the following format:

```javascript
{
  'A': [
    { to: 'B', distance: 10, cost: 5 },
    { to: 'C', distance: 15, cost: 3 }
  ],
  'B': [
    { to: 'D', distance: 5, cost: 7 }
  ],
  'C': [],
  'D': []
}
```

Each edge has two metrics:

- `distance`: the primary metric to minimize
- `cost`: the secondary metric limited by a budget constraint

Implement the following functions:

### `isValidGraph(graph)`

Validate a graph:

- Every `to` node exists as a key in the graph
- Every `distance` and `cost` value is a non-negative number
- No self-loops are allowed

Example:

```javascript
const valid = {
  'A': [{ to: 'B', distance: 5, cost: 3 }],
  'B': []
};
isValidGraph(valid); // => true

const invalid = {
  'A': [{ to: 'Z', distance: 5, cost: 3 }]
};
isValidGraph(invalid); // => false
```

### `findConstrainedPath(graph, start, end, maxCost)`

Find the shortest path by total distance while staying within the cost budget.

- The sum of all `cost` values on the path must not exceed `maxCost`
- Among all valid paths within budget, choose the one with the smallest total distance
- Cycles are not allowed
- Return `null` if no valid path exists

Example:

```javascript
const graph = {
  'A': [
    { to: 'B', distance: 5, cost: 100 },
    { to: 'C', distance: 10, cost: 5 }
  ],
  'B': [{ to: 'D', distance: 1, cost: 1 }],
  'C': [{ to: 'D', distance: 1, cost: 1 }],
  'D': []
};

findConstrainedPath(graph, 'A', 'D', 50);
// => { path: ['A', 'C', 'D'], totalDistance: 11, totalCost: 6 }

findConstrainedPath(graph, 'A', 'D', 150);
// => { path: ['A', 'B', 'D'], totalDistance: 6, totalCost: 101 }
```

Implementation notes:

- A modified Dijkstra algorithm is recommended
- Useful state representations include `(node, remainingBudget)` or `(node, costUsed)`
- Use a priority queue ordered by distance
- Prune dominated states when the same node is reached with higher cost

### `findKShortestPaths(graph, start, end, maxCost, k)`

Find the `k` shortest paths that satisfy the budget constraint.

- Every returned path must satisfy `maxCost`
- Return results sorted by ascending total distance
- Paths must be distinct
- If fewer than `k` valid paths exist, return all available paths

Example:

```javascript
findKShortestPaths(graph, 'A', 'D', 100, 3);
// => [
//   { path: ['A', 'B', 'D'], totalDistance: 10, totalCost: 15 },
//   { path: ['A', 'C', 'D'], totalDistance: 12, totalCost: 20 },
//   { path: ['A', 'X', 'Y', 'D'], totalDistance: 15, totalCost: 25 }
// ]
```

This is algorithmically challenging. Yen's K-shortest path algorithm or A* variants may help.

### `findPathWithWaypoints(graph, start, waypoints, end, maxCost)`

Find the shortest path from `start` to `end` that visits all `waypoints` in the given order.

- Waypoints must be visited in order
- The total cost across all segments must not exceed `maxCost`
- Each segment between waypoints should be chosen optimally

Example:

```javascript
findPathWithWaypoints(graph, 'A', ['B', 'C'], 'D', 50);
// Search for: A -> ... -> B -> ... -> C -> ... -> D
```

Implementation note:

- Break the problem into segments:
  `start -> waypoints[0]`, `waypoints[0] -> waypoints[1]`, ..., `waypoints[n] -> end`
- Use `findConstrainedPath` as a subroutine

## Context

In many real-world scenarios, paths must be optimized under more than one constraint. Common examples include:

- Route planning with toll budgets
- Network routing with latency and bandwidth costs
- Logistics with combined time and cost optimization

This challenge is intended to exercise graph algorithms, constrained search, path validation, and edge-case handling.

## Dependencies

- Node.js 18 or newer
- The included `package.json` uses Jest for tests

Local commands:

```bash
npm install
npm test
```

For watch mode:

```bash
npm run test:watch
```

## Constraints

- Do not change the provided public API
- Do not modify the tests
- Preserve the graph input format used by the tests
- Avoid cycles in returned paths
- Return `null` or `[]` exactly where required by the function contract

Expected complexity targets:

- `isValidGraph`: `O(V + E)`
- `findConstrainedPath`: `O((V × maxCost) × log(V × maxCost))` in the worst case
- `findKShortestPaths`: `O(k × V × (E + V log V))` or better
- `findPathWithWaypoints`: `O(|waypoints| × complexity(findConstrainedPath))`

## Edge Cases

- Empty graph
- `start === end`
- Missing target nodes in edge references
- Negative `distance` or `cost` values
- Self-loops
- No valid path within budget
- Fewer than `k` available valid paths
- Unreachable waypoints
- Budget exhaustion across waypoint segments
