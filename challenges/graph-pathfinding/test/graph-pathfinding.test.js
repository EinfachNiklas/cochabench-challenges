'use strict';

const {
	findConstrainedPath,
	findKShortestPaths,
	isValidGraph,
	findPathWithWaypoints
} = require('../graph-pathfinding');

describe('isValidGraph', () => {
	test('recognizes valid graph', () => {
		const graph = {
			'A': [{ to: 'B', distance: 5, cost: 10 }],
			'B': [{ to: 'C', distance: 3, cost: 5 }],
			'C': []
		};
		expect(isValidGraph(graph)).toBe(true);
	});

	test('rejects invalid "to" reference', () => {
		const graph = {
			'A': [{ to: 'Z', distance: 5, cost: 10 }],
			'B': []
		};
		expect(isValidGraph(graph)).toBe(false);
	});

	test('rejects negative values', () => {
		const graph = {
			'A': [{ to: 'B', distance: -5, cost: 10 }],
			'B': []
		};
		expect(isValidGraph(graph)).toBe(false);
	});

	test('rejects self-loops', () => {
		const graph = {
			'A': [{ to: 'A', distance: 1, cost: 1 }]
		};
		expect(isValidGraph(graph)).toBe(false);
	});

	test('accepts empty graph', () => {
		expect(isValidGraph({})).toBe(true);
	});
});

describe('findConstrainedPath', () => {
	test('finds simple direct path', () => {
		const graph = {
			'A': [{ to: 'B', distance: 10, cost: 5 }],
			'B': []
		};
		const result = findConstrainedPath(graph, 'A', 'B', 10);
		expect(result).toEqual({
			path: ['A', 'B'],
			totalDistance: 10,
			totalCost: 5
		});
	});

	test('returns null when budget is insufficient', () => {
		const graph = {
			'A': [{ to: 'B', distance: 10, cost: 100 }],
			'B': []
		};
		const result = findConstrainedPath(graph, 'A', 'B', 50);
		expect(result).toBe(null);
	});

	test('finds shortest path with multiple edges', () => {
		const graph = {
			'A': [
				{ to: 'B', distance: 10, cost: 20 },
				{ to: 'C', distance: 15, cost: 10 }
			],
			'B': [{ to: 'D', distance: 5, cost: 10 }],
			'C': [{ to: 'D', distance: 5, cost: 10 }],
			'D': []
		};
		const result = findConstrainedPath(graph, 'A', 'D', 30);
		// Shortest path: A->B->D (15) but costs 30
		// Alternative: A->C->D (20) costs 20
		expect(result.path).toEqual(['A', 'B', 'D']);
		expect(result.totalDistance).toBe(15);
		expect(result.totalCost).toBe(30);
	});

	test('chooses cheaper path when shortest is over budget', () => {
		const graph = {
			'A': [
				{ to: 'B', distance: 5, cost: 100 },
				{ to: 'C', distance: 10, cost: 5 }
			],
			'B': [{ to: 'D', distance: 1, cost: 1 }],
			'C': [{ to: 'D', distance: 1, cost: 1 }],
			'D': []
		};
		const result = findConstrainedPath(graph, 'A', 'D', 50);
		// A->B->D is shorter (6) but over budget (101)
		// A->C->D is longer (11) but within budget (6)
		expect(result.path).toEqual(['A', 'C', 'D']);
		expect(result.totalDistance).toBe(11);
		expect(result.totalCost).toBe(6);
	});

	test('returns null when no path exists', () => {
		const graph = {
			'A': [{ to: 'B', distance: 5, cost: 5 }],
			'B': [],
			'C': []
		};
		const result = findConstrainedPath(graph, 'A', 'C', 100);
		expect(result).toBe(null);
	});

	test('Start gleich Ziel', () => {
		const graph = { 'A': [] };
		const result = findConstrainedPath(graph, 'A', 'A', 100);
		expect(result).toEqual({
			path: ['A'],
			totalDistance: 0,
			totalCost: 0
		});
	});

	test('works with large complex graph', () => {
		const graph = {
			'A': [{ to: 'B', distance: 4, cost: 10 }, { to: 'C', distance: 2, cost: 3 }],
			'B': [{ to: 'D', distance: 5, cost: 4 }, { to: 'E', distance: 10, cost: 2 }],
			'C': [{ to: 'B', distance: 1, cost: 2 }, { to: 'D', distance: 8, cost: 7 }],
			'D': [{ to: 'E', distance: 2, cost: 5 }, { to: 'F', distance: 6, cost: 1 }],
			'E': [{ to: 'F', distance: 3, cost: 8 }],
			'F': []
		};
		const result = findConstrainedPath(graph, 'A', 'F', 15);
		expect(result).not.toBe(null);
		expect(result.path[0]).toBe('A');
		expect(result.path[result.path.length - 1]).toBe('F');
		expect(result.totalCost).toBeLessThanOrEqual(15);
	});
});

describe('findKShortestPaths', () => {
	test('finds multiple alternative paths', () => {
		const graph = {
			'A': [
				{ to: 'B', distance: 10, cost: 5 },
				{ to: 'C', distance: 15, cost: 3 }
			],
			'B': [{ to: 'D', distance: 5, cost: 5 }],
			'C': [{ to: 'D', distance: 5, cost: 5 }],
			'D': []
		};
		const results = findKShortestPaths(graph, 'A', 'D', 20, 2);
		expect(results).toHaveLength(2);
		expect(results[0].totalDistance).toBe(15); // A->B->D
		expect(results[1].totalDistance).toBe(20); // A->C->D
		expect(results[0].totalDistance).toBeLessThan(results[1].totalDistance);
	});

	test('returns empty array when no path exists', () => {
		const graph = {
			'A': [],
			'B': []
		};
		const results = findKShortestPaths(graph, 'A', 'B', 100, 3);
		expect(results).toEqual([]);
	});

	test('returns fewer than k paths when not enough exist', () => {
		const graph = {
			'A': [{ to: 'B', distance: 5, cost: 5 }],
			'B': []
		};
		const results = findKShortestPaths(graph, 'A', 'B', 10, 5);
		expect(results.length).toBeLessThanOrEqual(5);
		expect(results.length).toBeGreaterThan(0);
	});

	test('respects budget constraint', () => {
		const graph = {
			'A': [
				{ to: 'B', distance: 5, cost: 50 },
				{ to: 'C', distance: 10, cost: 5 }
			],
			'B': [],
			'C': []
		};
		const results = findKShortestPaths(graph, 'A', 'B', 20, 2);
		// A->B costs 50, over budget
		expect(results).toEqual([]);
	});

	test('sorts results by distance', () => {
		const graph = {
			'A': [
				{ to: 'B', distance: 20, cost: 1 },
				{ to: 'C', distance: 5, cost: 1 }
			],
			'B': [{ to: 'D', distance: 1, cost: 1 }],
			'C': [{ to: 'D', distance: 20, cost: 1 }],
			'D': []
		};
		const results = findKShortestPaths(graph, 'A', 'D', 10, 3);
		for (let i = 1; i < results.length; i++) {
			expect(results[i].totalDistance).toBeGreaterThanOrEqual(
				results[i - 1].totalDistance
			);
		}
	});
});

describe('findPathWithWaypoints', () => {
	test('finds path through multiple waypoints', () => {
		const graph = {
			'A': [{ to: 'B', distance: 5, cost: 5 }],
			'B': [{ to: 'C', distance: 5, cost: 5 }],
			'C': [{ to: 'D', distance: 5, cost: 5 }],
			'D': []
		};
		const result = findPathWithWaypoints(graph, 'A', ['B', 'C'], 'D', 20);
		expect(result).toEqual({
			path: ['A', 'B', 'C', 'D'],
			totalDistance: 15,
			totalCost: 15
		});
	});

	test('returns null when waypoint is unreachable', () => {
		const graph = {
			'A': [{ to: 'C', distance: 5, cost: 5 }],
			'B': [],
			'C': [{ to: 'D', distance: 5, cost: 5 }],
			'D': []
		};
		const result = findPathWithWaypoints(graph, 'A', ['B'], 'D', 50);
		expect(result).toBe(null);
	});

	test('respects budget across all segments', () => {
		const graph = {
			'A': [{ to: 'B', distance: 5, cost: 10 }],
			'B': [{ to: 'C', distance: 5, cost: 10 }],
			'C': [{ to: 'D', distance: 5, cost: 10 }],
			'D': []
		};
		const result = findPathWithWaypoints(graph, 'A', ['B', 'C'], 'D', 25);
		// Total: 30 cost, over budget
		expect(result).toBe(null);
	});

	test('works without waypoints', () => {
		const graph = {
			'A': [{ to: 'B', distance: 5, cost: 5 }],
			'B': []
		};
		const result = findPathWithWaypoints(graph, 'A', [], 'B', 10);
		expect(result).toEqual({
			path: ['A', 'B'],
			totalDistance: 5,
			totalCost: 5
		});
	});

	test('complex graph with alternative routes between waypoints', () => {
		const graph = {
			'A': [
				{ to: 'B', distance: 10, cost: 2 },
				{ to: 'X', distance: 5, cost: 1 }
			],
			'X': [{ to: 'B', distance: 5, cost: 1 }],
			'B': [
				{ to: 'C', distance: 10, cost: 2 },
				{ to: 'Y', distance: 5, cost: 1 }
			],
			'Y': [{ to: 'C', distance: 5, cost: 1 }],
			'C': []
		};
		const result = findPathWithWaypoints(graph, 'A', ['B'], 'C', 10);
		expect(result.path).toContain('B');
		expect(result.path[0]).toBe('A');
		expect(result.path[result.path.length - 1]).toBe('C');
		expect(result.totalCost).toBeLessThanOrEqual(10);
	});
});
